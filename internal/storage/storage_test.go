package storage

import (
	"fmt"
	"sync"
	"testing"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMemStorage_SetGauge(t *testing.T) {
	type fields struct {
		gauges   map[string]Gauge
		counters map[string]Counter
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
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
			}

			m.setGauge(tt.args.value, tt.args.name)
			assert.Equal(t, tt.wantFields, fields{m.gauges.data, m.counters.data})
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	type fields struct {
		gauges   map[string]Gauge
		counters map[string]Counter
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
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
			}

			m.addCounter(tt.args.value, tt.args.name)
			assert.Equal(t, tt.wantFields, fields{m.gauges.data, m.counters.data})
		})
	}
}

func TestMemStorage_GetGauge(t *testing.T) {
	type fields struct {
		gauges   map[string]Gauge
		counters map[string]Counter
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
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
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
		gauges   map[string]Gauge
		counters map[string]Counter
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
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
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
	type fields struct {
		gauges   map[string]Gauge
		counters map[string]Counter
	}
	type args struct {
		m models.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name: "add counter",
			fields: fields{
				gauges:   map[string]Gauge{"test": 0.1},
				counters: map[string]Counter{"test": 1},
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
				gauges:   map[string]Gauge{"test": 0.1},
				counters: map[string]Counter{"test": 1},
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
				gauges:   map[string]Gauge{"test": 0.1},
				counters: map[string]Counter{"test": 1},
			},
			args: args{
				m: models.Metrics{ID: "test", MType: "error"},
			},
			want:    models.Metrics{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
			}
			got, err := ms.UpdateByMetrics(tt.args.m)
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
		testGauge   float64 = 0.01
		testCounter int64   = 1
	)

	type fields struct {
		gauges   map[string]Gauge
		counters map[string]Counter
	}
	type args struct {
		m models.Metrics
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name:    "error gauge",
			fields:  fields{gauges: map[string]Gauge{"test": 0.01}},
			args:    args{models.Metrics{ID: "error", MType: "gauge"}},
			want:    models.Metrics{},
			wantErr: true,
		},
		{
			name:    "error counter",
			fields:  fields{counters: map[string]Counter{"test": 1}},
			args:    args{models.Metrics{ID: "error", MType: "counter"}},
			want:    models.Metrics{},
			wantErr: true,
		},
		{
			name:    "error type",
			fields:  fields{counters: map[string]Counter{"test": 1}},
			args:    args{models.Metrics{ID: "error", MType: "error"}},
			want:    models.Metrics{},
			wantErr: true,
		},
		{
			name:    "get gauge",
			fields:  fields{gauges: map[string]Gauge{"test": 0.01}},
			args:    args{models.Metrics{ID: "test", MType: "gauge"}},
			want:    models.Metrics{ID: "test", MType: "gauge", Value: &testGauge},
			wantErr: false,
		},
		{
			name:    "get counter",
			fields:  fields{counters: map[string]Counter{"test": 1}},
			args:    args{models.Metrics{ID: "test", MType: "counter"}},
			want:    models.Metrics{ID: "test", MType: "counter", Delta: &testCounter},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
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
	var (
		testCounter1 Counter = 1
		testCounter2 Counter = 2
	)
	type fields struct {
		gauges   map[string]Gauge
		counters map[string]Counter
	}
	type args struct {
		id    string
		delta *Counter
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name:    "test error",
			fields:  fields{counters: map[string]Counter{"test": 1}},
			args:    args{"test", nil},
			want:    models.Metrics{},
			wantErr: true,
		},
		{
			name:   fmt.Sprintf("test %d", testCounter1),
			fields: fields{counters: map[string]Counter{"test": 1}},
			args:   args{"test", &testCounter1},
			want:   models.NewMetricsForCounter("test", 2),
		},
		{
			name:   fmt.Sprintf("test %d", testCounter2),
			fields: fields{counters: map[string]Counter{"test": 1}},
			args:   args{"test", &testCounter2},
			want:   models.NewMetricsForCounter("test", 3),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
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
	var (
		testGauge1 Gauge = 1.1
		testGauge2 Gauge = 0
	)
	type fields struct {
		gauges   map[string]Gauge
		counters map[string]Counter
	}
	type args struct {
		id    string
		value *Gauge
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name:    "test error",
			fields:  fields{gauges: map[string]Gauge{"test": 0.01}},
			args:    args{"test", nil},
			want:    models.Metrics{},
			wantErr: true,
		},
		{
			name:   fmt.Sprintf("test %f", testGauge1),
			fields: fields{gauges: map[string]Gauge{"test": 0.01}},
			args:   args{"test", &testGauge1},
			want:   models.NewMetricsForGauge("test", 1.1),
		},
		{
			name:   fmt.Sprintf("test %f", testGauge2),
			fields: fields{gauges: map[string]Gauge{"test": 0.01}},
			args:   args{"test", &testGauge2},
			want:   models.NewMetricsForGauge("test", 0),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
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
		gauges   map[string]Gauge
		counters map[string]Counter
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name:    "not found",
			fields:  fields{counters: map[string]Counter{"test": 1}},
			args:    args{"error"},
			want:    models.Metrics{},
			wantErr: true,
		},
		{
			name:   "get value",
			fields: fields{counters: map[string]Counter{"test": 1}},
			args:   args{"test"},
			want:   models.NewMetricsForCounter("test", 1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
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
		gauges   map[string]Gauge
		counters map[string]Counter
	}
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    models.Metrics
		wantErr bool
	}{
		{
			name:    "not found",
			fields:  fields{gauges: map[string]Gauge{"test": 0.01}},
			args:    args{"error"},
			want:    models.Metrics{},
			wantErr: true,
		},
		{
			name:   "get value",
			fields: fields{gauges: map[string]Gauge{"test": 0.01}},
			args:   args{"test"},
			want:   models.NewMetricsForGauge("test", 0.01),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ms := &MemStorage{
				gauges: struct {
					sync.RWMutex
					data map[string]Gauge
				}{data: tt.fields.gauges},
				counters: struct {
					sync.RWMutex
					data map[string]Counter
				}{data: tt.fields.counters},
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
