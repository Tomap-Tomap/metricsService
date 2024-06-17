// Package ip describes the structures for working with IP
package ip

import (
	"context"
	"net"
	"net/http"
	"sync"

	"github.com/DarkOmap/metricsService/internal/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	headerXRealIP = "X-Real-IP"
)

var (
	localIP string
	once    sync.Once
)

// GetLocalIP returns the local IP
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

// InterceptorAddRealIP an interceptor for adding an X-Real-IP header
func InterceptorAddRealIP(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
	ctx = metadata.AppendToOutgoingContext(
		ctx,
		headerXRealIP, GetLocalIP(),
	)

	return invoker(ctx, method, req, reply, cc, opts...)
}

// Checker structure with methods for IP validation
type Checker struct {
	ipNet *net.IPNet
}

// NewChecker create IPChecker
func NewChecker(ipNet *net.IPNet) *Checker {
	return &Checker{ipNet: ipNet}
}

// RequsetIPCheck middleware checking IP
func (ipc *Checker) RequsetIPCheck(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		if ipc.ipNet == nil {
			next.ServeHTTP(w, r)
			return
		}

		realIP := r.Header.Get(headerXRealIP)

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

// InterceptorIPCheck interceptor checking IP
func (ipc *Checker) InterceptorIPCheck(ctx context.Context, req any, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	md, ok := metadata.FromIncomingContext(ctx)

	if !ok {
		err = status.Error(codes.PermissionDenied, "missing metadata")
		return
	}

	ip := md.Get(headerXRealIP)

	if len(ip) == 0 {
		err = status.Error(codes.PermissionDenied, "missing X-Real-IP")
		return
	}

	if !ipc.ipNet.Contains(net.ParseIP(ip[0])) {
		err = status.Error(codes.PermissionDenied, "network doesn't include given IP")
		return
	}

	resp, err = handler(ctx, req)

	return
}
