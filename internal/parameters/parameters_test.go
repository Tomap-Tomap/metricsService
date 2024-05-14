package parameters

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseFlagsAgent(t *testing.T) {
	tests := []struct {
		f                  func()
		name               string
		wantListenAddr     string
		wantCKP            string
		wantKey            string
		wantReportInterval uint
		wantPollInterval   uint
		wantRL             uint
	}{
		{
			name:               "test env",
			f:                  setEnv,
			wantListenAddr:     "testEnv",
			wantCKP:            "testPath",
			wantReportInterval: 10,
			wantPollInterval:   10,
			wantKey:            "key",
			wantRL:             5,
		},
		{
			name:               "test flags",
			f:                  setFlags,
			wantListenAddr:     "testFlags",
			wantCKP:            "testPath",
			wantReportInterval: 100,
			wantPollInterval:   100,
			wantKey:            "key",
			wantRL:             5,
		},
		{
			name:               "test default",
			f:                  nil,
			wantListenAddr:     "localhost:8080",
			wantCKP:            "",
			wantReportInterval: 10,
			wantPollInterval:   2,
			wantKey:            "",
			wantRL:             10,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.f != nil {
				tt.f()
			}

			p := ParseFlagsAgent()

			assert.Equal(t, tt.wantListenAddr, p.ListenAddr)
			assert.Equal(t, tt.wantReportInterval, p.ReportInterval)
			assert.Equal(t, tt.wantPollInterval, p.PollInterval)
			assert.Equal(t, tt.wantCKP, p.CryptoKeyPath)
			assert.Equal(t, tt.wantKey, p.Key)
			assert.Equal(t, tt.wantRL, p.RateLimit)

			delParameters()
		})
	}
}

func setEnv() {
	os.Setenv("ADDRESS", "testEnv")
	os.Setenv("CRYPTO_KEY", "testPath")
	os.Setenv("REPORT_INTERVAL", "10")
	os.Setenv("POLL_INTERVAL", "10")
	os.Setenv("KEY", "key")
	os.Setenv("RATE_LIMIT", "5")
}

func setFlags() {
	os.Args = []string{"test", "-a=testFlags", "-crypto-key=testPath", "-r=100", "-p=100", "-k=key", "-l=5"}
}

func delParameters() {
	os.Clearenv()
	os.Args = []string{"test"}
}

func TestParseFlagsServer(t *testing.T) {
	t.Run("test env", func(t *testing.T) {
		wantP := setEnvForServer()
		p := ParseFlagsServer()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test flags", func(t *testing.T) {
		wantP := setFlagsForServer()
		p := ParseFlagsServer()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test default", func(t *testing.T) {
		wantP := getDefaultParametersForServer()
		p := ParseFlagsServer()

		assert.Equal(t, wantP, p)
		delParameters()
	})
}

func setEnvForServer() ServerParameters {
	sp := ServerParameters{
		FlagRunAddr:     "testEnv",
		FileStoragePath: "/tmp/test.json",
		CryptoKeyPath:   "testPath",
		DataBaseDSN:     "test",
		StoreInterval:   10,
		Restore:         true,
		Key:             "key",
		RateLimit:       5,
	}
	os.Setenv("ADDRESS", sp.FlagRunAddr)
	os.Setenv("FILE_STORAGE_PATH", sp.FileStoragePath)
	os.Setenv("CRYPTO_KEY", sp.CryptoKeyPath)
	os.Setenv("DATABASE_DSN", sp.DataBaseDSN)
	os.Setenv("STORE_INTERVAL", "10")
	os.Setenv("RESTORE", "true")
	os.Setenv("KEY", sp.Key)
	os.Setenv("RATE_LIMIT", "5")

	return sp
}

func setFlagsForServer() ServerParameters {
	os.Args = []string{
		"test",
		"-a=testFlags",
		"-f=/tmp/test/test.json",
		"-crypto-key=testPath",
		"-d=testdb",
		"-i=10",
		"-r=false",
		"-k=key",
		"-l=5",
	}

	return ServerParameters{
		FlagRunAddr:     "testFlags",
		FileStoragePath: "/tmp/test/test.json",
		CryptoKeyPath:   "testPath",
		DataBaseDSN:     "testdb",
		StoreInterval:   10,
		Restore:         false,
		Key:             "key",
		RateLimit:       5,
	}
}

func getDefaultParametersForServer() ServerParameters {
	return ServerParameters{
		FlagRunAddr:     "localhost:8080",
		FileStoragePath: "/tmp/metrics-db.json",
		CryptoKeyPath:   "",
		DataBaseDSN:     "",
		StoreInterval:   300,
		Restore:         true,
		Key:             "",
		RateLimit:       10,
	}
}
