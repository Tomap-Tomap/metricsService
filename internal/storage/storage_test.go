package storage

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_memStorage_AddUnit(t *testing.T) {
	tests := []struct {
		name         string
		args         StorageUnit
		wantErr      bool
		wantGauges   map[string]gauge
		wantCounters map[string]counter
	}{
		{
			name:    "test error gauge",
			args:    StorageUnit{"gauge", "test", "test"},
			wantErr: true,
		},
		{
			name:    "test error counter",
			args:    StorageUnit{"counter", "test", "test"},
			wantErr: true,
		},
		{
			name:    "test error not type",
			args:    StorageUnit{"test", "test", "test"},
			wantErr: true,
		},
		{
			name:       "test add gauge",
			args:       StorageUnit{"gauge", "test", "11"},
			wantErr:    false,
			wantGauges: map[string]gauge{"test": gauge(11)},
		},
		{
			name:         "test add counter",
			args:         StorageUnit{"counter", "test", "11"},
			wantErr:      false,
			wantCounters: map[string]counter{"test": counter(11)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := memStorage{}
			m.counters = make(map[string]counter)
			m.gauges = make(map[string]gauge)

			err := m.AddUnit(tt.args)

			if !tt.wantErr {
				require.NoError(t, err)

				switch tt.args.unitType {
				case "gauge":
					assert.Subset(t, m.gauges, tt.wantGauges)
				case "counter":
					assert.Subset(t, m.counters, tt.wantCounters)
				}

				return
			}

			assert.Error(t, err)
		})
	}
}

func TestNewStorageUnit(t *testing.T) {
	type args struct {
		url string
	}
	tests := []struct {
		name    string
		args    args
		want    StorageUnit
		wantErr bool
	}{
		{
			name:    "test error",
			args:    args{"/gauge/test"},
			wantErr: true,
		},
		{
			name: "test gauge",
			args: args{"/gauge/test/12.5"},
			want: StorageUnit{"gauge", "test", "12.5"},
		},
		{
			name: "test counter",
			args: args{"/counter/test/12"},
			want: StorageUnit{"counter", "test", "12"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewStorageUnit(tt.args.url)

			if !tt.wantErr {
				require.NoError(t, err)

				assert.Equal(t, tt.want, got)

				return
			}

			assert.Error(t, err)
		})
	}
}
