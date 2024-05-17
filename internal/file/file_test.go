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

func TestProducer_WriteInFile(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		p, err := NewProducer("./testdata/testdb")
		require.NoError(t, err)

		want := map[string]string{"test": "test"}
		err = p.WriteInFile(want)
		require.NoError(t, err)
		p.Close()

		d, err := NewConsumer("./testdata/testdb")
		require.NoError(t, err)
		var got map[string]string
		err = d.Decoder.Decode(&got)
		require.NoError(t, err)
		d.Close()
		p, err = NewProducer("./testdata/testdb")
		require.NoError(t, err)
		p.ClearFile()
		p.Close()
		require.Equal(t, want, got)
	})

	t.Run("error test", func(t *testing.T) {
		p, err := NewProducer("./testdata/testdb")
		require.NoError(t, err)

		type Dummy struct {
			Name string
			Next *Dummy
		}
		dummy := Dummy{Name: "Dummy"}
		dummy.Next = &dummy

		err = p.WriteInFile(dummy)
		require.Error(t, err)
	})
}

func TestNewConsumer(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		c, err := NewConsumer("./testdata/testdb")
		require.NoError(t, err)
		c.Close()
	})

	t.Run("error test", func(t *testing.T) {
		_, err := NewConsumer("./testdata//")
		require.Error(t, err)
	})
}
