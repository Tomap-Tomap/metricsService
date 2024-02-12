package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/jackc/pgx/v5"
)

type CloseFunc func()

type DBStorage struct {
	conn         *pgx.Conn
	queryTimeout int64
}

func NewDBStorage(ctx context.Context, dsn string, queryTimeout int64) (*DBStorage, CloseFunc, error) {
	conn, err := pgx.Connect(ctx, dsn)

	if err != nil {
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	dbs := &DBStorage{conn, queryTimeout}

	if err := dbs.createTables(); err != nil {
		return nil, nil, fmt.Errorf("create tables in database: %w", err)
	}

	return dbs, func() { conn.Close(ctx) }, nil
}

func (dbs *DBStorage) createTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()

	selectQuery := `
		SELECT
			COUNT(table_name) > 0 AS tableExist
		FROM information_schema.tables
		WHERE
		table_name = $1
	`
	createGaugesQuery := `
		CREATE TABLE gauges (
			Id VARCHAR(20) PRIMARY KEY,
			Value DOUBLE PRECISION
		)
	`
	createCountersQuery := `
		CREATE TABLE counters (
			Id VARCHAR(20) PRIMARY KEY,
			Delta INTEGER
		)
	`
	var te bool
	err := dbs.conn.QueryRow(ctx, selectQuery, "gauges").Scan(&te)

	if err != nil {
		return fmt.Errorf("get gauges table: %w", err)
	}

	if !te {
		_, err := dbs.conn.Exec(ctx, createGaugesQuery)

		if err != nil {
			return fmt.Errorf("create gauges table: %w", err)
		}
	}

	te = false

	err = dbs.conn.QueryRow(ctx, selectQuery, "counters").Scan(&te)

	if err != nil {
		return fmt.Errorf("get counters table: %w", err)
	}

	if !te {
		_, err := dbs.conn.Exec(ctx, createCountersQuery)

		if err != nil {
			return fmt.Errorf("create counters table: %w", err)
		}
	}

	return nil
}

func (dbs *DBStorage) PingDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()
	return dbs.conn.Ping(ctx)
}

func (dbs *DBStorage) UpdateByMetrics(m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case "counter":
		return dbs.updateCounterByMetrics(m.ID, (*Counter)(m.Delta))
	case "gauge":
		return dbs.updateGaugeByMetrics(m.ID, (*Gauge)(m.Value))
	default:
		return nil, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (dbs *DBStorage) updateCounterByMetrics(id string, delta *Counter) (*models.Metrics, error) {
	if delta == nil {
		return nil, fmt.Errorf("delta is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()

	query := `
		WITH t AS (
			INSERT INTO counters (Id, Delta) VALUES ($1, $2)
			ON CONFLICT (Id) DO UPDATE SET Delta = counters.Delta + EXCLUDED.Delta
			RETURNING *
		)
		SELECT Delta FROM t WHERE Id = $1
	`

	var newDelta int64

	err := pgx.BeginFunc(ctx, dbs.conn, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, query, id, *delta).Scan(&newDelta)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("query execution: %w", err)
	}

	return models.NewMetricsForCounter(id, newDelta), nil
}

func (dbs *DBStorage) updateGaugeByMetrics(id string, value *Gauge) (*models.Metrics, error) {
	if value == nil {
		return nil, fmt.Errorf("value is empty")
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()

	query := `
		WITH t AS (
			INSERT INTO gauges (Id, Value) VALUES ($1, $2)
			ON CONFLICT (Id) DO UPDATE SET Value = EXCLUDED.Value
			RETURNING *
		)
		SELECT Value FROM t WHERE Id = $1
	`

	var newValue float64

	err := pgx.BeginFunc(ctx, dbs.conn, func(tx pgx.Tx) error {
		err := tx.QueryRow(ctx, query, id, *value).Scan(&newValue)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("query execution: %w", err)
	}

	return models.NewMetricsForGauge(id, newValue), nil
}

func (dbs *DBStorage) ValueByMetrics(m models.Metrics) (*models.Metrics, error) {
	switch m.MType {
	case "counter":
		return dbs.valueCounterByMetrics(m.ID)
	case "gauge":
		return dbs.valueGaugeByMetrics(m.ID)
	default:
		return nil, fmt.Errorf("unknown type %s", m.MType)
	}
}

func (dbs *DBStorage) valueCounterByMetrics(id string) (*models.Metrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()

	var c int64

	err := dbs.conn.QueryRow(ctx, "SELECT Delta FROM counters WHERE Id = $1", id).Scan(&c)

	if err != nil {
		return nil, fmt.Errorf("get counter %s: %w", id, err)
	}

	return models.NewMetricsForCounter(id, c), nil
}

func (dbs *DBStorage) valueGaugeByMetrics(id string) (*models.Metrics, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()

	var g float64

	err := dbs.conn.QueryRow(ctx, "SELECT Value FROM gauges WHERE Id = $1", id).Scan(&g)

	if err != nil {
		return nil, fmt.Errorf("get gauge %s: %w", id, err)
	}

	return models.NewMetricsForGauge(id, g), nil
}

func (dbs *DBStorage) GetAllGauge() map[string]Gauge {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()

	var (
		s      string
		g      float64
		retMap map[string]Gauge = make(map[string]Gauge)
	)

	rows, err := dbs.conn.Query(ctx, "SELECT Id, Value FROM gauges")

	if err != nil {
		return nil
	}

	_, err = pgx.ForEachRow(rows, []any{&s, &g}, func() error {
		retMap[s] = Gauge(g)
		return nil
	})

	if err != nil {
		return nil
	}

	return retMap
}

func (dbs *DBStorage) GetAllCounter() map[string]Counter {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(dbs.queryTimeout)*time.Second)
	defer cancel()

	var (
		s      string
		c      int64
		retMap map[string]Counter = make(map[string]Counter)
	)

	rows, err := dbs.conn.Query(ctx, "SELECT Id, Delta FROM counters")

	if err != nil {
		return nil
	}

	_, err = pgx.ForEachRow(rows, []any{&s, &c}, func() error {
		retMap[s] = Counter(c)
		return nil
	})

	if err != nil {
		return nil
	}

	return retMap
}
