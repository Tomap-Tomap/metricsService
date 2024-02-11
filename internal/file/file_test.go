package file

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewProducer(t *testing.T) {
	defer os.Remove("./test")
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name:    "test error",
			args:    args{"/"},
			wantErr: true,
		},
		{
			name:    "test no error",
			args:    args{"./test"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewProducer(tt.args.fileName)
			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}

}
