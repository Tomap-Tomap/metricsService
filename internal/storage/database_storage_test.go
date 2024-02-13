package storage

import (
	"context"
	"errors"
	"testing"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v3"
	"github.com/stretchr/testify/require"
)

func TestDBStorage_createTables(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	t.Run("test all table is create", func(t *testing.T) {
		mock.ExpectBegin()
		rows1 := pgxmock.NewRows([]string{"tableExist"}).AddRow(true)
		rows2 := pgxmock.NewRows([]string{"tableExist"}).AddRow(true)
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnRows(rows1)
		mock.ExpectQuery("SELECT").WithArgs("counters").WillReturnRows(rows2)
		mock.ExpectCommit()
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.NoError(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("test all table is not create", func(t *testing.T) {
		mock.ExpectBegin()
		rows1 := pgxmock.NewRows([]string{"tableExist"}).AddRow(false)
		rows2 := pgxmock.NewRows([]string{"tableExist"}).AddRow(false)
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnRows(rows1)
		mock.ExpectExec("CREATE TABLE gauges").WillReturnResult(pgconn.NewCommandTag("test"))
		mock.ExpectQuery("SELECT").WithArgs("counters").WillReturnRows(rows2)
		mock.ExpectExec("CREATE TABLE counters").WillReturnResult(pgconn.NewCommandTag("Test"))
		mock.ExpectCommit()
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.NoError(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("test gauges create", func(t *testing.T) {
		mock.ExpectBegin()
		rows1 := pgxmock.NewRows([]string{"tableExist"}).AddRow(true)
		rows2 := pgxmock.NewRows([]string{"tableExist"}).AddRow(false)
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnRows(rows1)
		mock.ExpectQuery("SELECT").WithArgs("counters").WillReturnRows(rows2)
		mock.ExpectExec("CREATE TABLE counters").WillReturnResult(pgconn.NewCommandTag("Test"))
		mock.ExpectCommit()
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.NoError(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("test counters create", func(t *testing.T) {
		mock.ExpectBegin()
		rows1 := pgxmock.NewRows([]string{"tableExist"}).AddRow(false)
		rows2 := pgxmock.NewRows([]string{"tableExist"}).AddRow(true)
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnRows(rows1)
		mock.ExpectExec("CREATE TABLE gauges").WillReturnResult(pgconn.NewCommandTag("Test"))
		mock.ExpectQuery("SELECT").WithArgs("counters").WillReturnRows(rows2)
		mock.ExpectCommit()
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.NoError(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error get gauges", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnError(errors.New("test"))
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.Error(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error gauges create", func(t *testing.T) {
		mock.ExpectBegin()
		rows1 := pgxmock.NewRows([]string{"tableExist"}).AddRow(false)
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnRows(rows1)
		mock.ExpectExec("CREATE TABLE gauges").WillReturnError(errors.New("test"))
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.Error(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error get counters", func(t *testing.T) {
		mock.ExpectBegin()
		rows1 := pgxmock.NewRows([]string{"tableExist"}).AddRow(true)
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnRows(rows1)
		mock.ExpectQuery("SELECT").WithArgs("counters").WillReturnError(errors.New("test"))
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.Error(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("erroe counters create", func(t *testing.T) {
		mock.ExpectBegin()
		rows1 := pgxmock.NewRows([]string{"tableExist"}).AddRow(true)
		rows2 := pgxmock.NewRows([]string{"tableExist"}).AddRow(false)
		mock.ExpectQuery("SELECT").WithArgs("gauges").WillReturnRows(rows1)
		mock.ExpectQuery("SELECT").WithArgs("counters").WillReturnRows(rows2)
		mock.ExpectExec("CREATE TABLE counters").WillReturnError(errors.New("test"))
		mock.ExpectRollback()

		dbs := DBStorage{conn: mock}
		require.Error(t, dbs.createTables())
		require.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestDBStorage_PingDB(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectPing()

	dbs := DBStorage{conn: mock}
	require.NoError(t, dbs.PingDB(context.Background()))
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_UpdateByMetrics(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	t.Run("test gauge update", func(t *testing.T) {
		mock.ExpectQuery("WITH").WithArgs("test", Gauge(1)).WillReturnRows(pgxmock.NewRows([]string{"Value"}).AddRow(float64(1)))
		dbs := DBStorage{conn: mock}

		m, err := dbs.UpdateByMetrics(context.Background(), *models.NewMetricsForGauge("test", 1))

		require.NoError(t, err)
		require.Equal(t, models.NewMetricsForGauge("test", 1), m)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("test counter update", func(t *testing.T) {
		mock.ExpectQuery("WITH").WithArgs("test", Counter(1)).WillReturnRows(pgxmock.NewRows([]string{"Delta"}).AddRow(int64(1)))
		dbs := DBStorage{conn: mock}

		m, err := dbs.UpdateByMetrics(context.Background(), *models.NewMetricsForCounter("test", 1))

		require.NoError(t, err)
		require.Equal(t, models.NewMetricsForCounter("test", 1), m)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("test unknown type", func(t *testing.T) {
		dbs := DBStorage{conn: mock}

		_, err := dbs.UpdateByMetrics(context.Background(), models.Metrics{ID: "test", MType: "test"})

		require.Error(t, err)
	})
}

func TestDBStorage_ValueByMetrics(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	t.Run("test gauge value", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs("test").WillReturnRows(pgxmock.NewRows([]string{"Value"}).AddRow(float64(1)))
		dbs := DBStorage{conn: mock}

		mod, _ := models.NewMetrics("test", "gauge")
		m, err := dbs.ValueByMetrics(context.Background(), *mod)

		require.NoError(t, err)
		require.Equal(t, models.NewMetricsForGauge("test", 1), m)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("test counter value", func(t *testing.T) {
		mock.ExpectQuery("SELECT").WithArgs("test").WillReturnRows(pgxmock.NewRows([]string{"Delta"}).AddRow(int64(1)))
		dbs := DBStorage{conn: mock}

		mod, _ := models.NewMetrics("test", "counter")
		m, err := dbs.ValueByMetrics(context.Background(), *mod)

		require.NoError(t, err)
		require.Equal(t, models.NewMetricsForCounter("test", 1), m)
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("test unknown type", func(t *testing.T) {
		dbs := DBStorage{conn: mock}

		_, err := dbs.ValueByMetrics(context.Background(), models.Metrics{ID: "test", MType: "test"})

		require.Error(t, err)
	})
}

func TestDBStorage_GetAllGauge(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectQuery("SELECT").
		WillReturnRows(pgxmock.NewRows([]string{"Name", "Value"}).
			AddRows([]any{"test1", float64(1)}, []any{"test2", float64(2)}))
	dbs := DBStorage{conn: mock}

	m, err := dbs.GetAllGauge(context.Background())

	require.NoError(t, err)
	require.Equal(t, map[string]Gauge{"test1": 1, "test2": 2}, m)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_GetAllCounter(t *testing.T) {
	mock, err := pgxmock.NewConn()
	require.NoError(t, err)
	defer mock.Close(context.Background())

	mock.ExpectQuery("SELECT").
		WillReturnRows(pgxmock.NewRows([]string{"Name", "Delta"}).
			AddRows([]any{"test1", int64(1)}, []any{"test2", int64(2)}))
	dbs := DBStorage{conn: mock}

	m, err := dbs.GetAllCounter(context.Background())

	require.NoError(t, err)
	require.Equal(t, map[string]Counter{"test1": 1, "test2": 2}, m)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDBStorage_errorHanlder(t *testing.T) {
	t.Run("all error", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewConn()
		require.NoError(t, err)
		defer mock.Close(context.Background())
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})

		dbs := DBStorage{conn: mock, retryCount: 3, duration: 1, durationPolicy: 2}

		err = dbs.errorHanlder(func() error {
			_, err := dbs.conn.Query(context.Background(), "SELECT")
			return err
		})

		require.Error(t, err)
		require.True(t, pgerrcode.IsConnectionException(err.(*pgconn.PgError).Code))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("error not class 08", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewConn()
		require.NoError(t, err)
		defer mock.Close(context.Background())
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "02000"})

		dbs := DBStorage{conn: mock, retryCount: 3, duration: 1, durationPolicy: 2}

		err = dbs.errorHanlder(func() error {
			_, err := dbs.conn.Query(context.Background(), "SELECT")
			return err
		})

		require.Error(t, err)
		require.True(t, !pgerrcode.IsConnectionException(err.(*pgconn.PgError).Code))
		require.NoError(t, mock.ExpectationsWereMet())
	})

	t.Run("not error", func(t *testing.T) {
		t.Parallel()
		mock, err := pgxmock.NewConn()
		require.NoError(t, err)
		defer mock.Close(context.Background())
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnError(&pgconn.PgError{Code: "08000"})
		mock.ExpectQuery("SELECT").WillReturnRows(pgxmock.NewRows([]string{"test"}).AddRow("test"))

		dbs := DBStorage{conn: mock, retryCount: 3, duration: 1, durationPolicy: 2}

		err = dbs.errorHanlder(func() error {
			_, err := dbs.conn.Query(context.Background(), "SELECT")
			return err
		})

		require.NoError(t, err)
	})
}
