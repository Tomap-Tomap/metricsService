package parameters

import (
	"flag"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlagsAgent(t *testing.T) {
	t.Run("test flags", func(t *testing.T) {
		wantP := setFlagsForAgent()
		p := ParseFlagsAgent()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test env", func(t *testing.T) {
		wantP := setEnvForAgent()
		p := ParseFlagsAgent()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test default", func(t *testing.T) {
		wantP := getDefaultParametersForAgent()
		p := ParseFlagsAgent()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test config file on flags c", func(t *testing.T) {
		wantP := setFlagsForAgent()
		os.Args = append(os.Args, "-c=./testdata/agent_config_test.json")
		p := ParseFlagsAgent()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test config file on flags config", func(t *testing.T) {
		wantP := setFlagsForAgent()
		os.Args = append(os.Args, "-config=./testdata/agent_config_test.json")
		p := ParseFlagsAgent()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test config file on env", func(t *testing.T) {
		wantP := setEnvForAgent()
		os.Setenv("CONFIG", "./testdata/agent_config_test.json")
		p := ParseFlagsAgent()

		assert.Equal(t, wantP, p)
		delParameters()
	})
}

func setEnvForAgent() AgentParameters {
	os.Setenv("ADDRESS", "testEnv")
	os.Setenv("CRYPTO_KEY", "testPath")
	os.Setenv("REPORT_INTERVAL", "10")
	os.Setenv("POLL_INTERVAL", "10")
	os.Setenv("KEY", "key")
	os.Setenv("RATE_LIMIT", "5")

	return AgentParameters{
		ListenAddr:     "testEnv",
		CryptoKeyPath:  "testPath",
		Key:            "key",
		ReportInterval: 10,
		RateLimit:      5,
		PollInterval:   10,
	}
}

func Test_parseAgentFromFile(t *testing.T) {
	t.Run("test no config file", func(t *testing.T) {
		wantP := getDefaultParametersForAgent()
		defer delParameters()

		var p AgentParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.UintVar(&p.ReportInterval, "r", 10, "report interval")
		f.UintVar(&p.PollInterval, "p", 2, "poll interval")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseAgentFromFile(f, &p, "")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test empty config file", func(t *testing.T) {
		wantP := setFlagsForAgent()
		defer delParameters()

		var p AgentParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.UintVar(&p.ReportInterval, "r", 10, "report interval")
		f.UintVar(&p.PollInterval, "p", 2, "poll interval")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseAgentFromFile(f, &p, "./testdata/agent_config_empty_test.json")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test config file", func(t *testing.T) {
		wantP := AgentParameters{
			ListenAddr:     "configAddr",
			CryptoKeyPath:  "configCKey",
			Key:            "configKey",
			ReportInterval: 111,
			RateLimit:      333,
			PollInterval:   222,
		}

		var p AgentParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.UintVar(&p.ReportInterval, "r", 10, "report interval")
		f.UintVar(&p.PollInterval, "p", 2, "poll interval")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseAgentFromFile(f, &p, "./testdata/agent_config_test.json")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test invalid file", func(t *testing.T) {
		var p AgentParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.UintVar(&p.ReportInterval, "r", 10, "report interval")
		f.UintVar(&p.PollInterval, "p", 2, "poll interval")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseAgentFromFile(f, &p, "./testdata/config_invalid_test")
		require.Error(t, err)
	})
}

func setFlagsForAgent() AgentParameters {
	os.Args = []string{
		"test",
		"-a=testFlags",
		"-crypto-key=testPath",
		"-r=100",
		"-p=100",
		"-k=key",
		"-l=5",
	}

	return AgentParameters{
		ListenAddr:     "testFlags",
		CryptoKeyPath:  "testPath",
		Key:            "key",
		ReportInterval: 100,
		RateLimit:      5,
		PollInterval:   100,
	}
}

func getDefaultParametersForAgent() AgentParameters {
	return AgentParameters{
		ListenAddr:     "localhost:8080",
		CryptoKeyPath:  "",
		Key:            "",
		ReportInterval: 10,
		RateLimit:      10,
		PollInterval:   2,
	}
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

	t.Run("test config file on flags c", func(t *testing.T) {
		wantP := setFlagsForServer()
		os.Args = append(os.Args, "-c=./testdata/server_config_test.json")
		p := ParseFlagsServer()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test config file on flags config", func(t *testing.T) {
		wantP := setFlagsForServer()
		os.Args = append(os.Args, "-config=./testdata/server_config_test.json")
		p := ParseFlagsServer()

		assert.Equal(t, wantP, p)
		delParameters()
	})

	t.Run("test config file on env", func(t *testing.T) {
		wantP := setEnvForServer()
		os.Setenv("CONFIG", "./testdata/server_config_test.json")
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

func Test_parseServerFromFile(t *testing.T) {
	t.Run("test no config file", func(t *testing.T) {
		wantP := getDefaultParametersForServer()
		defer delParameters()

		var p ServerParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.StringVar(&p.FileStoragePath, "f", "/tmp/metrics-db.json", "path to save storage")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to private key")
		f.StringVar(
			&p.DataBaseDSN,
			"d",
			"",
			"connection string to database",
		)
		f.UintVar(&p.StoreInterval, "i", 300, "interval in seconds for save storage")
		f.BoolVar(&p.Restore, "r", true, "flag for upload storage from file")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseServerFromFile(f, &p, "")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test empty config file", func(t *testing.T) {
		wantP := setFlagsForServer()
		defer delParameters()

		var p ServerParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.StringVar(&p.FileStoragePath, "f", "/tmp/metrics-db.json", "path to save storage")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to private key")
		f.StringVar(
			&p.DataBaseDSN,
			"d",
			"",
			"connection string to database",
		)
		f.UintVar(&p.StoreInterval, "i", 300, "interval in seconds for save storage")
		f.BoolVar(&p.Restore, "r", true, "flag for upload storage from file")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseServerFromFile(f, &p, "./testdata/server_config_empty_test.json")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test config file", func(t *testing.T) {
		wantP := ServerParameters{
			FlagRunAddr:     "configAddr",
			FileStoragePath: "configFile",
			CryptoKeyPath:   "configCKey",
			DataBaseDSN:     "configDSN",
			Key:             "configKey",
			StoreInterval:   111,
			Restore:         true,
			RateLimit:       222,
		}

		var p ServerParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.StringVar(&p.FileStoragePath, "f", "/tmp/metrics-db.json", "path to save storage")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to private key")
		f.StringVar(
			&p.DataBaseDSN,
			"d",
			"",
			"connection string to database",
		)
		f.UintVar(&p.StoreInterval, "i", 300, "interval in seconds for save storage")
		f.BoolVar(&p.Restore, "r", true, "flag for upload storage from file")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseServerFromFile(f, &p, "./testdata/server_config_test.json")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test invalid file", func(t *testing.T) {
		var p ServerParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		f.StringVar(&p.Key, "k", "", "hash key")
		f.StringVar(&p.FileStoragePath, "f", "/tmp/metrics-db.json", "path to save storage")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to private key")
		f.StringVar(
			&p.DataBaseDSN,
			"d",
			"",
			"connection string to database",
		)
		f.UintVar(&p.StoreInterval, "i", 300, "interval in seconds for save storage")
		f.BoolVar(&p.Restore, "r", true, "flag for upload storage from file")
		f.UintVar(&p.RateLimit, "l", 10, "rate limit")

		f.Parse(os.Args[1:])

		err := parseServerFromFile(f, &p, "./testdata/config_invalid_test")
		require.Error(t, err)
	})
}
