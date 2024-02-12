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
	conn *pgx.Conn
}

func NewDBStorage(ctx context.Context, dsn string) (*DBStorage, CloseFunc, error) {
	conn, err := pgx.Connect(ctx, dsn)

	if err != nil {
		return nil, nil, fmt.Errorf("connect to database: %w", err)
	}

	dbs := &DBStorage{conn}

	if err := dbs.createTables(); err != nil {
		return nil, nil, fmt.Errorf("create tables in database: %w", err)
	}

	return dbs, func() { conn.Close(ctx) }, nil
}

func (dbs *DBStorage) createTables() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
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
			Id SERIAL PRIMARY KEY,
			Name VARCHAR(150) UNIQUE,
			Value DOUBLE PRECISION
		);
		CREATE UNIQUE INDEX gauge_idx ON gauges (Name);
	`
	createCountersQuery := `
		CREATE TABLE counters (
			Id SERIAL PRIMARY KEY,
			Name VARCHAR(150) UNIQUE,
			Delta INTEGER
		);
		CREATE UNIQUE INDEX counter_idx ON counters (Name);
	`

	err := pgx.BeginFunc(ctx, dbs.conn, func(tx pgx.Tx) error {
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

	err := dbs.conn.QueryRow(ctx, query, id, *delta).Scan(&newDelta)

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

	err := dbs.conn.QueryRow(ctx, query, id, *value).Scan(&newValue)

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

	err := dbs.conn.QueryRow(ctx, "SELECT Delta FROM counters WHERE Name = $1", id).Scan(&c)

	if err != nil {
		return nil, fmt.Errorf("get counter %s: %w", id, err)
	}

	return models.NewMetricsForCounter(id, c), nil
}

func (dbs *DBStorage) valueGaugeByMetrics(ctx context.Context, id string) (*models.Metrics, error) {
	var g float64

	err := dbs.conn.QueryRow(ctx, "SELECT Value FROM gauges WHERE Name = $1", id).Scan(&g)

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
	)

	rows, err := dbs.conn.Query(ctx, "SELECT Name, Value FROM gauges")

	if err != nil {
		return nil, fmt.Errorf("get gauges from db: %w", err)
	}
	defer rows.Close()

	_, err = pgx.ForEachRow(rows, []any{&s, &g}, func() error {
		retMap[s] = Gauge(g)
		return nil
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
	)

	rows, err := dbs.conn.Query(ctx, "SELECT Name, Delta FROM counters")

	if err != nil {
		return nil, fmt.Errorf("get counters from db: %w", err)
	}
	defer rows.Close()

	_, err = pgx.ForEachRow(rows, []any{&s, &c}, func() error {
		retMap[s] = Counter(c)
		return nil
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
		if err := dbs.conn.SendBatch(ctx, batch).Close(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("send batch: %w", err)
	}
	return nil
}
