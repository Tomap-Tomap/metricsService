package ip

import (
	"net"
	"net/http"
	"sync"

	"github.com/DarkOmap/metricsService/internal/logger"
	"go.uber.org/zap"
)

var (
	localIP string
	once    sync.Once
)

func GetLocalIP() string {
	once.Do(func() {
		addrs, err := net.InterfaceAddrs()
		if err != nil {
			logger.Log.Error("Get interface addresses", zap.Error(err))
			return
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ipnet.IP.To4() != nil {
					localIP = ipnet.IP.String()
					return
				}
			}
		}
	})

	return localIP
}

type IPChecker struct {
	ipNet *net.IPNet
}

func NewIPChecker(ipNet *net.IPNet) *IPChecker {
	return &IPChecker{ipNet: ipNet}
}

func (ipc *IPChecker) RequsetIPCheck(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if ipc.ipNet == nil {
			next.ServeHTTP(w, r)
			return
		}

		realIP := r.Header.Get("X-Real-IP")

		if realIP == "" {
			http.Error(w, "empty header", http.StatusForbidden)
			return
		}

		if !ipc.ipNet.Contains(net.ParseIP(realIP)) {
			http.Error(w, "network doesn't include given IP", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}
