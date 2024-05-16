// Package parameters defines structure's parameters for agent/servers work.
package parameters

import (
	"cmp"
	"encoding/json"
	"flag"
	"os"
	"strconv"

	"github.com/DarkOmap/metricsService/internal/logger"
	"go.uber.org/zap"
)

// AgentParameters contains parameters for agent.
type AgentParameters struct {
	ListenAddr     string `json:"address"`
	CryptoKeyPath  string `json:"crypto_key"`
	Key            string `json:"key"`
	ReportInterval uint   `json:"report_interval"`
	RateLimit      uint   `json:"rate_limit"`
	PollInterval   uint   `json:"poll_interval"`
}

// ParseFlagsAgent return agent's parameters from console or env.
func ParseFlagsAgent() (p AgentParameters) {
	var config string
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		config = envConfig
	}

	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
	f.StringVar(&p.CryptoKeyPath, "crypto-key", "", "path to public key")
	f.StringVar(&p.Key, "k", "", "hash key")
	f.UintVar(&p.ReportInterval, "r", 10, "report interval")
	f.UintVar(&p.PollInterval, "p", 2, "poll interval")
	f.UintVar(&p.RateLimit, "l", 10, "rate limit")

	if config == "" {
		f.StringVar(&config, "c", "config.json", "path to agent configuration")
		f.StringVar(&config, "config", "config.json", "path to agent configuration")
	}

	f.Parse(os.Args[1:])

	parseAgentFromFile(f, &p, config)

	if envCKP := os.Getenv("CRYPTO_KEY"); envCKP != "" {
		p.CryptoKeyPath = envCKP
	}

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		p.ListenAddr = envAddr
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		p.Key = envKey
	}

	if envRI := os.Getenv("REPORT_INTERVAL"); envRI != "" {
		intRI, err := strconv.ParseUint(envRI, 10, 32)

		if err == nil {
			p.ReportInterval = uint(intRI)
		}
	}

	if envPI := os.Getenv("POLL_INTERVAL"); envPI != "" {
		intPI, err := strconv.ParseUint(envPI, 10, 32)

		if err == nil {
			p.PollInterval = uint(intPI)
		}
	}

	if envRL := os.Getenv("RATE_LIMIT"); envRL != "" {
		intRL, err := strconv.ParseUint(envRL, 10, 32)

		if err == nil {
			p.RateLimit = uint(intRL)
		}
	}

	return
}

func parseAgentFromFile(f *flag.FlagSet, p *AgentParameters, config string) {
	if config == "" {
		return
	}

	var jsonP AgentParameters

	file, err := os.Open(config)

	if err != nil {
		logger.Log.Warn("Failed open config file. Config file will not read", zap.Error(err))
		return
	}

	jd := json.NewDecoder(file)
	err = jd.Decode(&jsonP)

	if err != nil {
		logger.Log.Warn("Failed decode config file. Config file will not read", zap.Error(err))
		return
	}

	if p.ListenAddr == f.Lookup("a").DefValue {
		p.ListenAddr = cmp.Or(jsonP.ListenAddr, p.ListenAddr)
	}

	if p.CryptoKeyPath == f.Lookup("crypto-key").DefValue {
		p.CryptoKeyPath = cmp.Or(jsonP.CryptoKeyPath, p.CryptoKeyPath)
	}

	if p.Key == f.Lookup("k").DefValue {
		p.Key = cmp.Or(jsonP.Key, p.Key)
	}

	ri, _ := strconv.ParseUint(f.Lookup("r").DefValue, 10, 64)
	if p.ReportInterval == uint(ri) {
		p.ReportInterval = cmp.Or(jsonP.ReportInterval, p.ReportInterval)
	}

	pi, _ := strconv.ParseUint(f.Lookup("p").DefValue, 10, 64)
	if p.PollInterval == uint(pi) {
		p.PollInterval = cmp.Or(jsonP.PollInterval, p.PollInterval)
	}

	rl, _ := strconv.ParseUint(f.Lookup("l").DefValue, 10, 64)
	if p.RateLimit == uint(rl) {
		p.RateLimit = cmp.Or(jsonP.RateLimit, p.RateLimit)
	}
}

// ServerParameters contains parameters for server.
type ServerParameters struct {
	FlagRunAddr     string `json:"address"`
	FileStoragePath string `json:"file_storage_path"`
	CryptoKeyPath   string `json:"crypto_key"`
	DataBaseDSN     string `json:"database_dsn"`
	Key             string `json:"key"`
	StoreInterval   uint   `json:"store_interval"`
	Restore         bool   `json:"restore"`
	RateLimit       uint   `json:"rate_limit"`
}

// ParseFlagsServer return server's parameters from console or env.
func ParseFlagsServer() (p ServerParameters) {
	var config string
	if envConfig := os.Getenv("CONFIG"); envConfig != "" {
		config = envConfig
	}

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

	if config == "" {
		f.StringVar(&config, "c", "config.json", "path to server configuration")
		f.StringVar(&config, "config", "config.json", "path to server configuration")
	}

	f.Parse(os.Args[1:])

	parseServerFromFile(f, &p, config)

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		p.FlagRunAddr = envAddr
	}

	if envKey := os.Getenv("KEY"); envKey != "" {
		p.Key = envKey
	}

	if envSP := os.Getenv("FILE_STORAGE_PATH"); envSP != "" {
		p.FileStoragePath = envSP
	}

	if envCKP := os.Getenv("CRYPTO_KEY"); envCKP != "" {
		p.CryptoKeyPath = envCKP
	}

	if envDB := os.Getenv("DATABASE_DSN"); envDB != "" {
		p.DataBaseDSN = envDB
	}

	if envSI := os.Getenv("STORE_INTERVAL"); envSI != "" {
		if unitSI, err := strconv.ParseUint(envSI, 10, 32); err == nil {
			p.StoreInterval = uint(unitSI)
		}
	}

	if envR := os.Getenv("RESTORE"); envR != "" {
		if boolR, err := strconv.ParseBool(envR); err == nil {
			p.Restore = boolR
		}
	}

	if envRL := os.Getenv("RATE_LIMIT"); envRL != "" {
		intRL, err := strconv.ParseUint(envRL, 10, 32)

		if err == nil {
			p.RateLimit = uint(intRL)
		}
	}

	return
}

func parseServerFromFile(f *flag.FlagSet, p *ServerParameters, config string) {
	if config == "" {
		return
	}

	var jsonP ServerParameters

	file, err := os.Open(config)

	if err != nil {
		logger.Log.Warn("Failed open config file. Config file will not read", zap.Error(err))
		return
	}

	jd := json.NewDecoder(file)
	err = jd.Decode(&jsonP)

	if err != nil {
		logger.Log.Warn("Failed decode config file. Config file will not read", zap.Error(err))
		return
	}

	if p.FlagRunAddr == f.Lookup("a").DefValue {
		p.FlagRunAddr = cmp.Or(jsonP.FlagRunAddr, p.FlagRunAddr)
	}

	if p.FileStoragePath == f.Lookup("f").DefValue {
		p.FileStoragePath = cmp.Or(jsonP.FileStoragePath, p.FileStoragePath)
	}

	if p.CryptoKeyPath == f.Lookup("crypto-key").DefValue {
		p.CryptoKeyPath = cmp.Or(jsonP.CryptoKeyPath, p.CryptoKeyPath)
	}

	if p.DataBaseDSN == f.Lookup("d").DefValue {
		p.DataBaseDSN = cmp.Or(jsonP.DataBaseDSN, p.DataBaseDSN)
	}

	if p.Key == f.Lookup("k").DefValue {
		p.Key = cmp.Or(jsonP.Key, p.Key)
	}

	si, _ := strconv.ParseUint(f.Lookup("i").DefValue, 10, 64)
	if p.StoreInterval == uint(si) {
		p.StoreInterval = cmp.Or(jsonP.StoreInterval, p.StoreInterval)
	}

	rl, _ := strconv.ParseUint(f.Lookup("l").DefValue, 10, 64)
	if p.RateLimit == uint(rl) {
		p.RateLimit = cmp.Or(jsonP.RateLimit, p.RateLimit)
	}

	restore, _ := strconv.ParseBool(f.Lookup("r").DefValue)
	if p.Restore == restore {
		p.Restore = cmp.Or(jsonP.Restore, p.Restore)
	}
}
