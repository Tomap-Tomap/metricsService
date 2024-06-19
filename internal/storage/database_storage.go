package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/parameters"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	queryUpdateGauges = `
		WITH t AS (
			INSERT INTO gauges (Name, Value) VALUES ($1, $2)
			ON CONFLICT (Name) DO UPDATE SET Value = EXCLUDED.Value
			RETURNING *
		)
		SELECT Value FROM t WHERE Name = $1
	`
	queryUpdateCounters = `
		WITH t AS (
			INSERT INTO counters (Name, Delta) VALUES ($1, $2)
			ON CONFLICT (Name) DO UPDATE SET Delta = counters.Delta + EXCLUDED.Delta
			RETURNING *
		)
		SELECT Delta FROM t WHERE Name = $1
	`
)

type retryPolicy struct {
	retryCount int
	duration   int
	increment  int
}

// DBStorage contains methods for working with postgres storage.
type DBStorage struct {
	conn        *pgxpool.Pool
	retryPolicy retryPolicy
}

// NewDBStorage create DBStorage
func NewDBStorage(ctx context.Context, p parameters.ServerParameters) (*DBStorage, error) {
	conn, err := pgxpool.New(ctx, p.DataBaseDSN)
	if err != nil {
		return nil, fmt.Errorf("create pgxpool: %w", err)
	}

	rp := retryPolicy{3, 1, 2}
	dbs := &DBStorage{conn: conn, retryPolicy: rp}

	if err := dbs.createTables(); err != nil {
		return nil, fmt.Errorf("create tables in database: %w", err)
	}

	return dbs, nil
}

// Close closes DBStorage
func (dbs *DBStorage) Close() error {
	dbs.conn.Close()

	return nil
}

// PingDB checks database.
func (dbs *DBStorage) PingDB(ctx context.Context) error {
	return dbs.conn.Ping(ctx)
}

// UpdateByMetrics updates data on database and returns new data.
func (dbs *DBStorage) UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case models.TypeCounter:
		return dbs.updateCounterByMetrics(ctx, m.ID, (*Counter)(m.Delta))
	case models.TypeGauge:
		return dbs.updateGaugeByMetrics(ctx, m.ID, (*Gauge)(m.Value))
	default:
		return nil, ErrUnknownType
	}
}

// ValueByMetrics returns data from database.
func (dbs *DBStorage) ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case models.TypeCounter:
		return dbs.valueCounterByMetrics(ctx, m.ID)
	case models.TypeGauge:
		return dbs.valueGaugeByMetrics(ctx, m.ID)
	default:
		return nil, ErrUnknownType
	}
}

// GetAll returns all data from DBStorage
func (dbs *DBStorage) GetAll(ctx context.Context) (map[string]fmt.Stringer, error) {
	var (
		s      string
		g      Gauge
		c      Counter
		t      string
		retMap = make(map[string]fmt.Stringer)
		rows   pgx.Rows
	)

	rows, err := retry2[pgx.Rows](ctx, dbs.retryPolicy, func() (pgx.Rows, error) {
		return dbs.conn.Query(ctx,
			`SELECT
				Name, Value as Value, 0 as Delta, 'gauge' as Type
			FROM gauges
			UNION
			SELECT
				 Name, 0, Delta, 'counter' as Type
			FROM counters`,
		)
	})
	if err != nil {
		return nil, fmt.Errorf("get data from db: %w", err)
	}
	defer rows.Close()

	_, err = retry2[pgconn.CommandTag](ctx, dbs.retryPolicy, func() (pgconn.CommandTag, error) {
		return pgx.ForEachRow(rows, []any{&s, &g, &c, &t}, func() error {
			if t == models.TypeGauge {
				retMap[s] = g
			} else {
				retMap[s] = c
			}

			return nil
		})
	})
	if err != nil {
		return nil, fmt.Errorf("parse data from db: %w", err)
	}

	return retMap, nil
}

// Updates updates database's datas.
func (dbs *DBStorage) Updates(ctx context.Context, metrics []models.Metrics) error {
	batch := &pgx.Batch{}

	for _, val := range metrics {
		switch val.MType {
		case models.TypeGauge:
			batch.Queue(queryUpdateGauges, val.ID, *val.Value)
		case models.TypeCounter:
			batch.Queue(queryUpdateCounters, val.ID, *val.Delta)
		}
	}

	err := pgx.BeginFunc(ctx, dbs.conn, func(tx pgx.Tx) error {
		err := retry(ctx, dbs.retryPolicy, func() error {
			return dbs.conn.SendBatch(ctx, batch).Close()
		})

		return err
	})
	if err != nil {
		return fmt.Errorf("send batch: %w", err)
	}
	return nil
}

func (dbs *DBStorage) createTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	createGaugesQuery := `
		CREATE TABLE IF NOT EXISTS gauges (
			Id SERIAL PRIMARY KEY,
			Name VARCHAR(150) UNIQUE,
			Value DOUBLE PRECISION
		);
		CREATE UNIQUE INDEX IF NOT EXISTS gauge_idx ON gauges (Name);
	`
	createCountersQuery := `
		CREATE TABLE IF NOT EXISTS counters (
			Id SERIAL PRIMARY KEY,
			Name VARCHAR(150) UNIQUE,
			Delta BIGINT
		);
		CREATE UNIQUE INDEX IF NOT EXISTS counter_idx ON counters (Name);
	`

	err := pgx.BeginFunc(ctx, dbs.conn, func(tx pgx.Tx) error {
		_, err := retry2[pgconn.CommandTag](ctx, dbs.retryPolicy, func() (pgconn.CommandTag, error) {
			return dbs.conn.Exec(ctx, createGaugesQuery)
		})
		if err != nil {
			return fmt.Errorf("create gauges table: %w", err)
		}

		_, err = retry2[pgconn.CommandTag](ctx, dbs.retryPolicy, func() (pgconn.CommandTag, error) {
			return dbs.conn.Exec(ctx, createCountersQuery)
		})
		if err != nil {
			return fmt.Errorf("create counters table: %w", err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (dbs *DBStorage) updateCounterByMetrics(ctx context.Context, id string, delta *Counter) (*models.Metrics, error) {
	if delta == nil {
		return nil, ErrEmptyDelta
	}

	var newDelta int64

	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, queryUpdateCounters, id, *delta).Scan(&newDelta)
	})
	if err != nil {
		return nil, fmt.Errorf("update counter metric name %s delta %d: %w", id, *delta, err)
	}

	return models.NewMetricsForCounter(id, newDelta), nil
}

func (dbs *DBStorage) updateGaugeByMetrics(ctx context.Context, id string, value *Gauge) (*models.Metrics, error) {
	if value == nil {
		return nil, ErrEmptyValue
	}

	var newValue float64

	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, queryUpdateGauges, id, *value).Scan(&newValue)
	})
	if err != nil {
		return nil, fmt.Errorf("update gauge metric name %s value %f: %w", id, *value, err)
	}

	return models.NewMetricsForGauge(id, newValue), nil
}

func (dbs *DBStorage) valueCounterByMetrics(ctx context.Context, id string) (*models.Metrics, error) {
	var c int64
	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, "SELECT Delta FROM counters WHERE Name = $1", id).Scan(&c)
	})
	if err != nil {
		return nil, fmt.Errorf("get counter in DB %s: %w", id, err)
	}

	return models.NewMetricsForCounter(id, c), nil
}

func (dbs *DBStorage) valueGaugeByMetrics(ctx context.Context, id string) (*models.Metrics, error) {
	var g float64
	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, "SELECT Value FROM gauges WHERE Name = $1", id).Scan(&g)
	})
	if err != nil {
		return nil, fmt.Errorf("get gauge in DB %s: %w", id, err)
	}

	return models.NewMetricsForGauge(id, g), nil
}

func retry(ctx context.Context, rp retryPolicy, fn func() error) error {
	fnWithReturn := func() (struct{}, error) {
		return struct{}{}, fn()
	}

	_, err := retry2(ctx, rp, fnWithReturn)
	return err
}

func retry2[T any](ctx context.Context, rp retryPolicy, fn func() (T, error)) (T, error) {
	if val1, err := fn(); err == nil || !isonnectionException(err) {
		return val1, err
	}

	var err error
	var ret1 T
	duration := rp.duration
	for i := 0; i < rp.retryCount; i++ {
		select {
		case <-time.NewTimer(time.Duration(duration) * time.Second).C:
			ret1, err = fn()
			if err == nil || !isonnectionException(err) {
				return ret1, err
			}
		case <-ctx.Done():
			return ret1, err
		}

		duration += rp.increment
	}

	return ret1, err
}

func isonnectionException(err error) bool {
	var tError *pgconn.PgError
	if errors.As(err, &tError) && pgerrcode.IsConnectionException(tError.Code) {
		return true
	}

	return false
}
