package storage

import (
	"fmt"
	"os"
	"testing"

	"github.com/DarkOmap/metricsService/internal/file"
	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_SetGauge(t *testing.T) {
	defer os.Remove("./test")
	producer, err := file.NewProducer("./test")

	require.NoError(t, err)
	defer producer.Close()

	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		value Gauge
		name  string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantFields fields
	}{
		{
			name:       "add gauge",
			fields:     fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:       args{1.11, "tg"},
			wantFields: fields{map[string]Gauge{"test": 0.12, "tg": 1.11}, map[string]Counter{"test": 1}},
		},
		{
			name:       "add gauge exchange",
			fields:     fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:       args{1.11, "test"},
			wantFields: fields{map[string]Gauge{"test": 1.11}, map[string]Counter{"test": 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: producer,
			}

			m.setGauge(tt.args.value, tt.args.name)
			assert.Equal(t, tt.wantFields, fields{m.Gauges.Data, m.Counters.Data})
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	defer os.Remove("./test")
	producer, err := file.NewProducer("./test")

	require.NoError(t, err)
	defer producer.Close()

	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		value Counter
		name  string
	}
	tests := []struct {
		name       string
		fields     fields
		args       args
		wantErr    bool
		wantFields fields
	}{
		{
			name:       "add counter",
			fields:     fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:       args{1, "tc"},
			wantFields: fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1, "tc": 1}},
		},
		{
			name:       "add counter increment",
			fields:     fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:       args{1, "test"},
			wantFields: fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 2}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: producer,
			}

			m.addCounter(tt.args.value, tt.args.name)
			assert.Equal(t, tt.wantFields, fields{m.Gauges.Data, m.Counters.Data})
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	tests := []struct {
		name    string
		fields  fields
		args    string
		want    Gauge
		wantErr bool
	}{
		{
			name:    "gauge not found",
			fields:  fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:    "wrongName",
			wantErr: true,
		},
		{
			name:   "get gauge",
			fields: fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:   "test",
			want:   Gauge(0.12),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: &file.Producer{},
			}
			got, err := m.getGauge(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetCounter(t *testing.T) {
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	tests := []struct {
		name    string
		fields  fields
		args    string
		want    Counter
		wantErr bool
	}{
		{
			name:    "counter not found",
			fields:  fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:    "wrongName",
			wantErr: true,
		},
		{
			name:   "get counter",
			fields: fields{map[string]Gauge{"test": 0.12}, map[string]Counter{"test": 1}},
			args:   "test",
			want:   Counter(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: &file.Producer{},
			}
			got, err := m.getCounter(tt.args)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_UpdateByMetrics(t *testing.T) {
	defer os.Remove("./test")
	producer, err := file.NewProducer("./test")

	require.NoError(t, err)
	defer producer.Close()

	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		m *models.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Metrics
		wantErr bool
	}{
		{
			name: "add counter",
			fields: fields{
				Gauges:   map[string]Gauge{"test": 0.1},
				Counters: map[string]Counter{"test": 1},
			},
			args: args{
				m: models.NewMetricsForCounter("test", 1),
			},
			want:    models.NewMetricsForCounter("test", 2),
			wantErr: false,
		},
		{
			name: "set gauge",
			fields: fields{
				Gauges:   map[string]Gauge{"test": 0.1},
				Counters: map[string]Counter{"test": 1},
			},
			args: args{
				m: models.NewMetricsForGauge("test", 1.1),
			},
			want:    models.NewMetricsForGauge("test", 1.1),
			wantErr: false,
		},
		{
			name: "unknown type",
			fields: fields{
				Gauges:   map[string]Gauge{"test": 0.1},
				Counters: map[string]Counter{"test": 1},
			},
			args: args{
				m: &models.Metrics{ID: "test", MType: "error"},
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: producer,
			}
			got, err := ms.UpdateByMetrics(*tt.args.m)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_ValueByMetrics(t *testing.T) {
	var (
		testGauge         = 0.01
		testCounter int64 = 1
	)

	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		m models.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Metrics
		wantErr bool
	}{
		{
			name:    "error gauge",
			fields:  fields{Gauges: map[string]Gauge{"test": 0.01}},
			args:    args{models.Metrics{ID: "error", MType: "gauge"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error counter",
			fields:  fields{Counters: map[string]Counter{"test": 1}},
			args:    args{models.Metrics{ID: "error", MType: "counter"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "error type",
			fields:  fields{Counters: map[string]Counter{"test": 1}},
			args:    args{models.Metrics{ID: "error", MType: "error"}},
			want:    nil,
			wantErr: true,
		},
		{
			name:    "get gauge",
			fields:  fields{Gauges: map[string]Gauge{"test": 0.01}},
			args:    args{models.Metrics{ID: "test", MType: "gauge"}},
			want:    &models.Metrics{ID: "test", MType: "gauge", Value: &testGauge},
			wantErr: false,
		},
		{
			name:    "get counter",
			fields:  fields{Counters: map[string]Counter{"test": 1}},
			args:    args{models.Metrics{ID: "test", MType: "counter"}},
			want:    &models.Metrics{ID: "test", MType: "counter", Delta: &testCounter},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: &file.Producer{},
			}
			got, err := ms.ValueByMetrics(tt.args.m)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_updateCounterByMetrics(t *testing.T) {
	defer os.Remove("./test")
	producer, err := file.NewProducer("./test")

	require.NoError(t, err)
	defer producer.Close()

	var (
		testCounter1 Counter = 1
		testCounter2 Counter = 2
	)
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		id    string
		delta *Counter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Metrics
		wantErr bool
	}{
		{
			name:    "test error",
			fields:  fields{Counters: map[string]Counter{"test": 1}},
			args:    args{"test", nil},
			want:    nil,
			wantErr: true,
		},
		{
			name:   fmt.Sprintf("test %d", testCounter1),
			fields: fields{Counters: map[string]Counter{"test": 1}},
			args:   args{"test", &testCounter1},
			want:   models.NewMetricsForCounter("test", 2),
		},
		{
			name:   fmt.Sprintf("test %d", testCounter2),
			fields: fields{Counters: map[string]Counter{"test": 1}},
			args:   args{"test", &testCounter2},
			want:   models.NewMetricsForCounter("test", 3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: producer,
			}
			got, err := ms.updateCounterByMetrics(tt.args.id, tt.args.delta)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_updateGaugeByMetrics(t *testing.T) {
	defer os.Remove("./test")
	producer, err := file.NewProducer("./test")

	require.NoError(t, err)
	defer producer.Close()
	var (
		testGauge1 Gauge = 1.1
		testGauge2 Gauge = 0
	)
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		id    string
		value *Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Metrics
		wantErr bool
	}{
		{
			name:    "test error",
			fields:  fields{Gauges: map[string]Gauge{"test": 0.01}},
			args:    args{"test", nil},
			want:    nil,
			wantErr: true,
		},
		{
			name:   fmt.Sprintf("test %f", testGauge1),
			fields: fields{Gauges: map[string]Gauge{"test": 0.01}},
			args:   args{"test", &testGauge1},
			want:   models.NewMetricsForGauge("test", 1.1),
		},
		{
			name:   fmt.Sprintf("test %f", testGauge2),
			fields: fields{Gauges: map[string]Gauge{"test": 0.01}},
			args:   args{"test", &testGauge2},
			want:   models.NewMetricsForGauge("test", 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: producer,
			}
			got, err := ms.updateGaugeByMetrics(tt.args.id, tt.args.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_valueCounter(t *testing.T) {
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Metrics
		wantErr bool
	}{
		{
			name:    "not found",
			fields:  fields{Counters: map[string]Counter{"test": 1}},
			args:    args{"error"},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "get value",
			fields: fields{Counters: map[string]Counter{"test": 1}},
			args:   args{"test"},
			want:   models.NewMetricsForCounter("test", 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: &file.Producer{},
			}
			got, err := ms.valueCounterByMetrics(tt.args.id)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_valueGaugeByMetrics(t *testing.T) {
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *models.Metrics
		wantErr bool
	}{
		{
			name:    "not found",
			fields:  fields{Gauges: map[string]Gauge{"test": 0.01}},
			args:    args{"error"},
			want:    nil,
			wantErr: true,
		},
		{
			name:   "get value",
			fields: fields{Gauges: map[string]Gauge{"test": 0.01}},
			args:   args{"test"},
			want:   models.NewMetricsForGauge("test", 0.01),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: &file.Producer{},
			}
			got, err := ms.valueGaugeByMetrics(tt.args.id)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMemStorage_GetAllGauge(t *testing.T) {
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	tests := []struct {
		name       string
		fields     fields
		wantRetMap map[string]Gauge
	}{
		{
			name:       "get all Gauges",
			fields:     fields{Gauges: map[string]Gauge{"test1": 1, "test2": 2, "test3": 3}},
			wantRetMap: map[string]Gauge{"test1": 1, "test2": 2, "test3": 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: &file.Producer{},
			}
			gotRetMap := ms.GetAllGauge()
			assert.Equal(t, tt.wantRetMap, gotRetMap)
		})
	}
}

func TestMemStorage_GetAllCounter(t *testing.T) {
	type fields struct {
		Gauges   map[string]Gauge
		Counters map[string]Counter
	}
	tests := []struct {
		name       string
		fields     fields
		wantRetMap map[string]Counter
	}{
		{
			name:       "get all caounters",
			fields:     fields{Counters: map[string]Counter{"test1": 1, "test2": 2, "test3": 3}},
			wantRetMap: map[string]Counter{"test1": 1, "test2": 2, "test3": 3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				Gauges: gauges{
					Data: tt.fields.Gauges,
				},
				Counters: counters{
					Data: tt.fields.Counters,
				},
				producer: &file.Producer{},
			}
			gotRetMap := ms.GetAllCounter()
			assert.Equal(t, tt.wantRetMap, gotRetMap)
		})
	}
}
