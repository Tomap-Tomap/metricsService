package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBStorage struct {
	conn        *pgx.Conn
	retryPolicy retryPolicy
}

type retryPolicy struct {
	retryCount int
	duration   int
	increment  int
}

func NewDBStorage(conn *pgx.Conn) (*DBStorage, error) {
	rp := retryPolicy{3, 1, 2}
	dbs := &DBStorage{conn: conn, retryPolicy: rp}

	if err := dbs.createTables(); err != nil {
		return nil, fmt.Errorf("create tables in database: %w", err)
	}

	return dbs, nil
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

func (dbs *DBStorage) PingDB(ctx context.Context) error {
	return dbs.conn.Ping(ctx)
}

func (dbs *DBStorage) UpdateByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case "counter":
		return dbs.updateCounterByMetrics(ctx, m.ID, (*Counter)(m.Delta))
	case "gauge":
		return dbs.updateGaugeByMetrics(ctx, m.ID, (*Gauge)(m.Value))
	default:
		return nil, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (dbs *DBStorage) updateCounterByMetrics(ctx context.Context, id string, delta *Counter) (*models.Metrics, error) {
	if delta == nil {
		return nil, fmt.Errorf("delta is empty")
	}

	query := `
		WITH t AS (
			INSERT INTO counters (Name, Delta) VALUES ($1, $2)
			ON CONFLICT (Name) DO UPDATE SET Delta = counters.Delta + EXCLUDED.Delta
			RETURNING *
		)
		SELECT Delta FROM t WHERE Name = $1
	`

	var newDelta int64

	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, query, id, *delta).Scan(&newDelta)
	})

	if err != nil {
		return nil, fmt.Errorf("query execution: %w", err)
	}

	return models.NewMetricsForCounter(id, newDelta), nil
}

func (dbs *DBStorage) updateGaugeByMetrics(ctx context.Context, id string, value *Gauge) (*models.Metrics, error) {
	if value == nil {
		return nil, fmt.Errorf("value is empty")
	}

	query := `
		WITH t AS (
			INSERT INTO gauges (Name, Value) VALUES ($1, $2)
			ON CONFLICT (Name) DO UPDATE SET Value = EXCLUDED.Value
			RETURNING *
		)
		SELECT Value FROM t WHERE Name = $1
	`

	var newValue float64

	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, query, id, *value).Scan(&newValue)
	})

	if err != nil {
		return nil, fmt.Errorf("query execution: %w", err)
	}

	return models.NewMetricsForGauge(id, newValue), nil
}

func (dbs *DBStorage) ValueByMetrics(ctx context.Context, m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case "counter":
		return dbs.valueCounterByMetrics(ctx, m.ID)
	case "gauge":
		return dbs.valueGaugeByMetrics(ctx, m.ID)
	default:
		return nil, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (dbs *DBStorage) valueCounterByMetrics(ctx context.Context, id string) (*models.Metrics, error) {
	var c int64
	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, "SELECT Delta FROM counters WHERE Name = $1", id).Scan(&c)
	})

	if err != nil {
		return nil, fmt.Errorf("get counter %s: %w", id, err)
	}

	return models.NewMetricsForCounter(id, c), nil
}

func (dbs *DBStorage) valueGaugeByMetrics(ctx context.Context, id string) (*models.Metrics, error) {
	var g float64
	err := retry(ctx, dbs.retryPolicy, func() error {
		return dbs.conn.QueryRow(ctx, "SELECT Value FROM gauges WHERE Name = $1", id).Scan(&g)
	})

	if err != nil {
		return nil, fmt.Errorf("get gauge %s: %w", id, err)
	}

	return models.NewMetricsForGauge(id, g), nil
}

func (dbs *DBStorage) GetAllGauge(ctx context.Context) (map[string]Gauge, error) {
	var (
		s      string
		g      float64
		retMap = make(map[string]Gauge)
		rows   pgx.Rows
	)

	rows, err := retry2[pgx.Rows](ctx, dbs.retryPolicy, func() (pgx.Rows, error) {
		return dbs.conn.Query(ctx, "SELECT Name, Value FROM gauges")
	})

	if err != nil {
		return nil, fmt.Errorf("get gauges from db: %w", err)
	}
	defer rows.Close()

	_, err = retry2[pgconn.CommandTag](ctx, dbs.retryPolicy, func() (pgconn.CommandTag, error) {
		return pgx.ForEachRow(rows, []any{&s, &g}, func() error {
			retMap[s] = Gauge(g)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("parse gauges from db: %w", err)
	}

	return retMap, nil
}

func (dbs *DBStorage) GetAllCounter(ctx context.Context) (map[string]Counter, error) {
	var (
		s      string
		c      int64
		retMap = make(map[string]Counter)
		rows   pgx.Rows
	)

	rows, err := retry2[pgx.Rows](ctx, dbs.retryPolicy, func() (pgx.Rows, error) {
		return dbs.conn.Query(ctx, "SELECT Name, Delta FROM counters")
	})

	if err != nil {
		return nil, fmt.Errorf("get counters from db: %w", err)
	}
	defer rows.Close()

	_, err = retry2[pgconn.CommandTag](ctx, dbs.retryPolicy, func() (pgconn.CommandTag, error) {
		return pgx.ForEachRow(rows, []any{&s, &c}, func() error {
			retMap[s] = Counter(c)
			return nil
		})
	})

	if err != nil {
		return nil, fmt.Errorf("parse counters from db: %w", err)
	}

	return retMap, nil
}

func (dbs *DBStorage) Updates(ctx context.Context, metrics []models.Metrics) error {
	batch := &pgx.Batch{}

	queryGauges := `
		WITH t AS (
			INSERT INTO gauges (Name, Value) VALUES ($1, $2)
			ON CONFLICT (Name) DO UPDATE SET Value = EXCLUDED.Value
			RETURNING *
		)
		SELECT Value FROM t WHERE Name = $1
	`

	queryCounters := `
		WITH t AS (
			INSERT INTO counters (Name, Delta) VALUES ($1, $2)
			ON CONFLICT (Name) DO UPDATE SET Delta = counters.Delta + EXCLUDED.Delta
			RETURNING *
		)
		SELECT Delta FROM t WHERE Name = $1
	`

	for _, val := range metrics {
		switch val.MType {
		case "gauge":
			batch.Queue(queryGauges, val.ID, *val.Value)
		case "counter":
			batch.Queue(queryCounters, val.ID, *val.Delta)
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
