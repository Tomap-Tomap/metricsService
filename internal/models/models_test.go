package models

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseGauge(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    float64
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
			got, err := parseGauge(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_parseCounter(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    int64
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
			got, err := parseCounter(tt.arg)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMetrics(t *testing.T) {
	type args struct {
		id    string
		mType string
	}
	tests := []struct {
		name    string
		args    args
		want    *Metrics
		wantErr bool
	}{
		{
			name:    fmt.Sprintf("test id %s mType %s", "test", "error"),
			args:    args{id: "test", mType: "error"},
			want:    nil,
			wantErr: true,
		},
		{
			name: fmt.Sprintf("test id %s mType %s", "test", "gauge"),
			args: args{id: "test", mType: "gauge"},
			want: &Metrics{ID: "test", MType: "gauge"},
		},
		{
			name: fmt.Sprintf("test id %s mType %s", "test", "counter"),
			args: args{id: "test", mType: "counter"},
			want: &Metrics{ID: "test", MType: "counter"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewMetrics(tt.args.id, tt.args.mType)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewModelByURL(t *testing.T) {
	var (
		testGauge         = 1.1
		testCounter int64 = 1
	)
	type args struct {
		name      string
		valueType string
		value     string
	}
	tests := []struct {
		name    string
		args    args
		want    *Metrics
		wantErr bool
	}{
		{
			name:    "test gauge",
			args:    args{"test", "gauge", "1.1"},
			want:    &Metrics{ID: "test", MType: "gauge", Value: &testGauge},
			wantErr: false,
		},
		{
			name:    "test counter",
			args:    args{"test", "counter", "1"},
			want:    &Metrics{ID: "test", MType: "counter", Delta: &testCounter},
			wantErr: false,
		},
		{
			name:    "test error",
			args:    args{"test", "test", "1"},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewModelByStrings(tt.args.name, tt.args.valueType, tt.args.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewModelsByJSON(t *testing.T) {
	var (
		testGauge         = 1.1
		testCounter int64 = 1
	)

	type args struct {
		j []byte
	}
	tests := []struct {
		name    string
		args    args
		want    *Metrics
		wantErr bool
	}{
		{
			name: "test gauge",
			args: args{[]byte(`{
					"id":"test",
					"type":"gauge",
					"value":1.1
				}`)},
			want:    &Metrics{ID: "test", MType: "gauge", Value: &testGauge},
			wantErr: false,
		},
		{
			name: "test counter",
			args: args{[]byte(`{
					"id":"test",
					"type":"counter",
					"delta":1
				}`)},
			want:    &Metrics{ID: "test", MType: "counter", Delta: &testCounter},
			wantErr: false,
		},
		{
			name: "test error json format",
			args: args{[]byte(`{
					"id":"test",
					"type":"counter",
					"delta":1,
				}`)},
			want:    nil,
			wantErr: true,
		},
		{
			name: "test unknown type",
			args: args{[]byte(`{
					"id":"test",
					"type":"error",
					"delta":1
				}`)},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewModelsByJSON(tt.args.j)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_checkType(t *testing.T) {
	type args struct {
		mType string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test gauge",
			args:    args{"gauge"},
			wantErr: false,
		},
		{
			name:    "test counter",
			args:    args{"counter"},
			wantErr: false,
		},
		{
			name:    "test error",
			args:    args{"error"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkType(tt.args.mType)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestNewMetricsForGauge(t *testing.T) {
	var (
		testGauge1         = 1.1
		testGauge2         = 0.1
		testGauge3 float64 = 0
	)
	type args struct {
		id    string
		value float64
	}
	tests := []struct {
		name string
		args args
		want *Metrics
	}{
		{
			name: fmt.Sprintf("test %f", testGauge1),
			args: args{"test", testGauge1},
			want: &Metrics{ID: "test", MType: "gauge", Value: &testGauge1},
		},
		{
			name: fmt.Sprintf("test %f", testGauge2),
			args: args{"test", testGauge2},
			want: &Metrics{ID: "test", MType: "gauge", Value: &testGauge2},
		},
		{
			name: fmt.Sprintf("test %f", testGauge3),
			args: args{"test", testGauge3},
			want: &Metrics{ID: "test", MType: "gauge", Value: &testGauge3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetricsForGauge(tt.args.id, tt.args.value)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNewMetricsForCounter(t *testing.T) {
	var (
		testCounter1 int64 = 1
		testCounter2 int64 = 2
		testCounter3 int64 = 0
	)
	type args struct {
		id    string
		delta int64
	}
	tests := []struct {
		name string
		args args
		want *Metrics
	}{
		{
			name: fmt.Sprintf("test %d", testCounter1),
			args: args{"test", testCounter1},
			want: &Metrics{ID: "test", MType: "counter", Delta: &testCounter1},
		},
		{
			name: fmt.Sprintf("test %d", testCounter2),
			args: args{"test", testCounter2},
			want: &Metrics{ID: "test", MType: "counter", Delta: &testCounter2},
		},
		{
			name: fmt.Sprintf("test %d", testCounter3),
			args: args{"test", testCounter3},
			want: &Metrics{ID: "test", MType: "counter", Delta: &testCounter3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewMetricsForCounter(tt.args.id, tt.args.delta)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_counterMetricsBySting(t *testing.T) {
	var (
		testCounter1 int64 = 1
		testCounter2 int64 = 2
		testCounter3 int64 = 0
	)
	type args struct {
		id    string
		delta string
	}
	tests := []struct {
		name    string
		args    args
		want    *Metrics
		wantErr bool
	}{
		{
			name:    "test error",
			args:    args{"error", "error"},
			want:    nil,
			wantErr: true,
		},
		{
			name: fmt.Sprintf("test %d", testCounter1),
			args: args{"test", "1"},
			want: &Metrics{ID: "test", MType: "counter", Delta: &testCounter1},
		},
		{
			name: fmt.Sprintf("test %d", testCounter2),
			args: args{"test", "2"},
			want: &Metrics{ID: "test", MType: "counter", Delta: &testCounter2},
		},
		{
			name: fmt.Sprintf("test %d", testCounter3),
			args: args{"test", "0"},
			want: &Metrics{ID: "test", MType: "counter", Delta: &testCounter3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := counterMetricsBySting(tt.args.id, tt.args.delta)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func Test_gaugeMetricsByStrings(t *testing.T) {
	var (
		testGauge1         = 1.1
		testGauge2         = 0.1
		testGauge3 float64 = 0
	)
	type args struct {
		id    string
		value string
	}
	tests := []struct {
		name    string
		args    args
		want    *Metrics
		wantErr bool
	}{
		{
			name:    "test error",
			args:    args{"error", "error"},
			want:    nil,
			wantErr: true,
		},
		{
			name: fmt.Sprintf("test %f", testGauge1),
			args: args{"test", "1.1"},
			want: &Metrics{ID: "test", MType: "gauge", Value: &testGauge1},
		},
		{
			name: fmt.Sprintf("test %f", testGauge2),
			args: args{"test", "0.1"},
			want: &Metrics{ID: "test", MType: "gauge", Value: &testGauge2},
		},
		{
			name: fmt.Sprintf("test %f", testGauge3),
			args: args{"test", "0"},
			want: &Metrics{ID: "test", MType: "gauge", Value: &testGauge3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gaugeMetricsByStrings(tt.args.id, tt.args.value)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
