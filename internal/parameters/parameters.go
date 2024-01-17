package parameters

import (
	"flag"
	"os"
	"strconv"
)

type AgentParameters struct {
	ListenAddr                   string
	ReportInterval, PollInterval uint
}

type ServerParameters struct {
	FlagRunAddr string
}

func ParseFlagsAgent() (p AgentParameters) {
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.StringVar(&p.ListenAddr, "a", "localhost:8080", "address and port to server")
	f.UintVar(&p.ReportInterval, "r", 10, "report interval")
	f.UintVar(&p.PollInterval, "p", 2, "poll interval")
	f.Parse(os.Args[1:])

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		p.ListenAddr = envAddr
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

	return
}

func ParseFlagsServer() (p ServerParameters) {
	f := flag.NewFlagSet(os.Args[0], flag.ContinueOnError)
	f.StringVar(&p.FlagRunAddr, "a", "localhost:8080", "address and port to run server")
	f.Parse(os.Args[1:])

	if envAddr := os.Getenv("ADDRESS"); envAddr != "" {
		p.FlagRunAddr = envAddr
	}

	return
}
