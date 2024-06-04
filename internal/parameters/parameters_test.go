package parameters

import (
	"flag"
	"net"
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
	os.Setenv("USE_GRPC", "True")

	return AgentParameters{
		ListenAddr:     "testEnv",
		CryptoKeyPath:  "testPath",
		HashKey:        "key",
		ReportInterval: 10,
		RateLimit:      5,
		PollInterval:   10,
		UseGRPC:        true,
	}
}

func Test_parseAgentFromFile(t *testing.T) {
	t.Run("test no config file", func(t *testing.T) {
		wantP := getDefaultParametersForAgent()
		defer delParameters()

		var p AgentParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
		f.BoolVar(&p.UseGRPC, "grpc", false, "flag for using grpc client, else use http client")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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
		f.BoolVar(&p.UseGRPC, "grpc", false, "flag for using grpc client, else use http client")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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
			HashKey:        "configKey",
			ReportInterval: 111,
			RateLimit:      333,
			PollInterval:   222,
			UseGRPC:        true,
		}

		var p AgentParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
		f.BoolVar(&p.UseGRPC, "grpc", false, "flag for using grpc client, else use http client")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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
		f.BoolVar(&p.UseGRPC, "grpc", false, "flag for using grpc client, else use http client")
		f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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
		"-grpc=true",
	}

	return AgentParameters{
		ListenAddr:     "testFlags",
		CryptoKeyPath:  "testPath",
		HashKey:        "key",
		ReportInterval: 100,
		RateLimit:      5,
		PollInterval:   100,
		UseGRPC:        true,
	}
}

func getDefaultParametersForAgent() AgentParameters {
	return AgentParameters{
		ListenAddr:     "localhost:8080",
		CryptoKeyPath:  "",
		HashKey:        "",
		ReportInterval: 10,
		RateLimit:      10,
		PollInterval:   2,
		UseGRPC:        false,
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
	_, ts, _ := net.ParseCIDR("192.168.1.0/24")

	sp := ServerParameters{
		FlagRunAddr:     "testEnv",
		FlagRunGRPCAddr: "testGRPCEnv",
		FileStoragePath: "/tmp/test.json",
		CryptoKeyPath:   "testPath",
		DataBaseDSN:     "test",
		StoreInterval:   10,
		Restore:         true,
		HashKey:         "key",
		RateLimit:       5,
		TrustedSubnet:   ts,
	}
	os.Setenv("ADDRESS", sp.FlagRunAddr)
	os.Setenv("GRPC_ADDRESS", sp.FlagRunGRPCAddr)
	os.Setenv("FILE_STORAGE_PATH", sp.FileStoragePath)
	os.Setenv("CRYPTO_KEY", sp.CryptoKeyPath)
	os.Setenv("DATABASE_DSN", sp.DataBaseDSN)
	os.Setenv("STORE_INTERVAL", "10")
	os.Setenv("RESTORE", "true")
	os.Setenv("KEY", sp.HashKey)
	os.Setenv("RATE_LIMIT", "5")
	os.Setenv("TRUSTED_SUBNET", "192.168.1.0/24")

	return sp
}

func setFlagsForServer() ServerParameters {
	os.Args = []string{
		"test",
		"-a=testFlags",
		"-grpc-a=testGRPCFlags",
		"-f=/tmp/test/test.json",
		"-crypto-key=testPath",
		"-d=testdb",
		"-i=10",
		"-r=false",
		"-k=key",
		"-l=5",
		"-t=192.168.1.0/24",
	}

	_, ts, _ := net.ParseCIDR("192.168.1.0/24")
	return ServerParameters{
		FlagRunAddr:     "testFlags",
		FlagRunGRPCAddr: "testGRPCFlags",
		FileStoragePath: "/tmp/test/test.json",
		CryptoKeyPath:   "testPath",
		DataBaseDSN:     "testdb",
		StoreInterval:   10,
		Restore:         false,
		HashKey:         "key",
		RateLimit:       5,
		TrustedSubnet:   ts,
	}
}

func getDefaultParametersForServer() ServerParameters {
	return ServerParameters{
		FlagRunAddr:     "localhost:8080",
		FlagRunGRPCAddr: "localhost:3200",
		FileStoragePath: "/tmp/metrics-db.json",
		CryptoKeyPath:   "",
		DataBaseDSN:     "",
		StoreInterval:   300,
		Restore:         true,
		HashKey:         "",
		RateLimit:       10,
		TrustedSubnet:   nil,
	}
}

func Test_parseServerFromFile(t *testing.T) {
	t.Run("test no config file", func(t *testing.T) {
		wantP := getDefaultParametersForServer()
		defer delParameters()

		var p ServerParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		f.StringVar(&p.FlagRunGRPCAddr, "grpc-a", "localhost:3200", "address and port to run server")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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
		f.StringVar(&p.FlagRunGRPCAddr, "grpc-a", "localhost:3200", "address and port to run server")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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

		var trustedSubnet string
		f.StringVar(&trustedSubnet, "t", "192.168.1.0/24", "trusted subnet")

		f.Parse(os.Args[1:])

		_, p.TrustedSubnet, _ = net.ParseCIDR(trustedSubnet)

		err := parseServerFromFile(f, &p, "./testdata/server_config_empty_test.json")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test config file", func(t *testing.T) {
		_, wantCIDR, err := net.ParseCIDR("192.168.1.0/24")

		require.NoError(t, err)

		wantP := ServerParameters{
			FlagRunAddr:     "configAddr",
			FlagRunGRPCAddr: "configGRPCAddr",
			FileStoragePath: "configFile",
			CryptoKeyPath:   "configCKey",
			DataBaseDSN:     "configDSN",
			HashKey:         "configKey",
			StoreInterval:   111,
			Restore:         true,
			RateLimit:       222,
			TrustedSubnet:   wantCIDR,
		}

		var p ServerParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		f.StringVar(&p.FlagRunGRPCAddr, "grpc-a", "localhost:3200", "address and port to run server")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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

		err = parseServerFromFile(f, &p, "./testdata/server_config_test.json")
		require.NoError(t, err)
		require.Equal(t, wantP, p)
	})

	t.Run("test invalid file", func(t *testing.T) {
		var p ServerParameters
		f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
		f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
		f.StringVar(&p.FlagRunGRPCAddr, "grpc-a", "localhost:3200", "address and port to run server")
		f.StringVar(&p.HashKey, "k", "", "hash key")
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
