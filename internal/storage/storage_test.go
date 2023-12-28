package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testTyper uint8

func (t testTyper) getType() int {
	return -1
}

func Test_memStorage_AddValue(t *testing.T) {
	type fields struct {
		gauges   map[string]gauge
		counters map[string]counter
	}
	type args struct {
		value Typer
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
			name:    "negative",
			fields:  fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:    args{testTyper(1), "negative"},
			wantErr: true,
		},
		{
			name:       "add counter",
			fields:     fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:       args{counter(1), "tc"},
			wantFields: fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1, "tc": 1}},
		},
		{
			name:       "add counter increment",
			fields:     fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:       args{counter(1), "test"},
			wantFields: fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 2}},
		},
		{
			name:       "add gauge",
			fields:     fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:       args{gauge(1.11), "tg"},
			wantFields: fields{map[string]gauge{"test": 0.12, "tg": 1.11}, map[string]counter{"test": 1}},
		},
		{
			name:       "add gauge exchange",
			fields:     fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:       args{gauge(1.11), "test"},
			wantFields: fields{map[string]gauge{"test": 1.11}, map[string]counter{"test": 1}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gauges:   tt.fields.gauges,
				counters: tt.fields.counters,
			}

			err := m.AddValue(tt.args.value, tt.args.name)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.wantFields, fields{m.gauges, m.counters})
		})
	}
}

func Test_memStorage_GetValue(t *testing.T) {
	type fields struct {
		gauges   map[string]gauge
		counters map[string]counter
	}
	type args struct {
		valueType int
		name      string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Typer
		wantErr bool
	}{
		{
			name:    "gauge not found",
			fields:  fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:    args{GaugeType, "wrongName"},
			wantErr: true,
		},
		{
			name:    "counter not found",
			fields:  fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:    args{CounterType, "wrongName"},
			wantErr: true,
		},
		{
			name:    "wrong type",
			fields:  fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:    args{-1, "wrongName"},
			wantErr: true,
		},
		{
			name:   "get gauge",
			fields: fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:   args{GaugeType, "test"},
			want:   gauge(0.12),
		},
		{
			name:   "get counter",
			fields: fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			args:   args{CounterType, "test"},
			want:   counter(1),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gauges:   tt.fields.gauges,
				counters: tt.fields.counters,
			}
			got, err := m.GetValue(tt.args.valueType, tt.args.name)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_memStorage_GetData(t *testing.T) {
	type fields struct {
		gauges   map[string]gauge
		counters map[string]counter
	}
	tests := []struct {
		name   string
		fields fields
		want   []dataResult
	}{
		{
			name:   "test full",
			fields: fields{map[string]gauge{"test": 0.12}, map[string]counter{"test": 1}},
			want:   []dataResult{{"test", counter(1)}, {"test", gauge(0.12)}},
		},
		{
			name:   "test only gauge",
			fields: fields{map[string]gauge{"test": 0.12}, nil},
			want:   []dataResult{{"test", gauge(0.12)}},
		},
		{
			name:   "test only counters",
			fields: fields{nil, map[string]counter{"test": 1}},
			want:   []dataResult{{"test", counter(1)}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				gauges:   tt.fields.gauges,
				counters: tt.fields.counters,
			}
			got := m.GetData()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseType(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    int
		wantErr bool
	}{
		{
			name:    "wrong type",
			arg:     "wrongType",
			wantErr: true,
		},
		{
			name: "gauge type",
			arg:  "gauge",
			want: GaugeType,
		},
		{
			name: "counter type",
			arg:  "counter",
			want: CounterType,
		},
		{
			name: "gauge type upper",
			arg:  "Gauge",
			want: GaugeType,
		},
		{
			name: "counter type upper",
			arg:  "Counter",
			want: CounterType,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseType(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseGauge(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    gauge
		wantErr bool
	}{
		{
			name:    "wront number",
			arg:     "test",
			wantErr: true,
		},
		{
			name: "test 0.001",
			arg:  "0.001",
			want: 0.001,
		},
		{
			name: "test 1.001",
			arg:  "1.001",
			want: 1.001,
		},
		{
			name: "test -1.001",
			arg:  "-1.001",
			want: -1.001,
		},
		{
			name: "test -0",
			arg:  "-0",
			want: -0,
		},
		{
			name: "test 0",
			arg:  "0",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseGauge(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestParseCounter(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    counter
		wantErr bool
	}{
		{
			name:    "wront number",
			arg:     "test",
			wantErr: true,
		},
		{
			name: "test 1",
			arg:  "1",
			want: 1,
		},
		{
			name: "test -1",
			arg:  "-1",
			want: -1,
		},
		{
			name: "test 11",
			arg:  "11",
			want: 11,
		},
		{
			name: "test -0",
			arg:  "-0",
			want: -0,
		},
		{
			name: "test 0",
			arg:  "0",
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCounter(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
