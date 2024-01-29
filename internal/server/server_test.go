package server

import (
	"testing"

	"github.com/DarkOmap/metricsService/internal/storage"
	"github.com/stretchr/testify/require"
)

func TestServer_initializeStoreFunc(t *testing.T) {
	type fields struct {
		fileStoragePath string
		fr              fileRepository
	}
	type args struct {
		storeInterval uint
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name:   "test interval 0",
			fields: fields{"test", storage.NewMemStorage()},
			args:   args{0},
		},
		{
			name:   "test interval not 0",
			fields: fields{"test", storage.NewMemStorage()},
			args:   args{30},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				fileStoragePath: tt.fields.fileStoragePath,
				fr:              tt.fields.fr,
			}

			err := s.initializeStoreFunc(tt.args.storeInterval)
			require.NoError(t, err)
			require.NotNil(t, s.storeFunc)
		})
	}
}
