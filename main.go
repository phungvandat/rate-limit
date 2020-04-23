package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/phungvandat/rate-limit/limiter"
)

// Rate limit
var visitors = limiter.NewExecutor([]limiter.Rule{
	limiter.Rule{
		Key:             "/",
		DurationSeconds: 60,
		Max:             10,
	},
})

func main() {
	err := godotenv.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load .env by error: %v", err))
	}
	port := 4000
	addr := fmt.Sprintf(":%v", port)

	serverCRT := os.Getenv("SERVER_CRT")
	serverKey := os.Getenv("SERVER_KEY")

	creds, err := tls.X509KeyPair([]byte(serverCRT), []byte(serverKey))
	if err != nil {
		panic(err)
	}
	configTLS := &tls.Config{
		Certificates: []tls.Certificate{creds},
	}
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		panic(err)
	}

	tlsListener := tls.NewListener(listener, configTLS)

	server := &http.Server{
		Handler: newHTTPHandler(),
	}
	errChn := make(chan error)
	go func() {
		fmt.Println("HTTPS: ", port)
		errChn <- server.Serve(tlsListener)
	}()

	go func() {
		sysChn := make(chan os.Signal)
		signal.Notify(sysChn, syscall.SIGINT, syscall.SIGTERM)
		errChn <- fmt.Errorf("%v", <-sysChn)
	}()

	go visitors.CleanupVisitors()
	log.Fatalf("Exit: %v", <-errChn)
}

func newHTTPHandler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", rateLimit(indexFunc))

	return mux
}

func rateLimit(next http.HandlerFunc) http.HandlerFunc {
	return func(res http.ResponseWriter, req *http.Request) {
		var (
			ipAddr = getRequestIP(req)
			path   = req.URL.Path
		)
		visitor := visitors.GetVisitor(ipAddr, path)
		if !visitor.Allow() {
			http.Error(res, "Rate limit error", http.StatusLocked)
			return
		}
		next(res, req)
	}
}

func indexFunc(res http.ResponseWriter, req *http.Request) {
	res.WriteHeader(200)
	res.Write([]byte("Success"))
}

// getRequestIP func
func getRequestIP(req *http.Request) string {
	ipAddress := req.Header.Get("X-Real-Ip")
	if ipAddress == "" {
		ipAddress = req.Header.Get("X-Forwarded-For")
	}
	if ipAddress == "" {
		stringArr := strings.Split(req.RemoteAddr, ":")
		if len(stringArr) > 0 {
			ipAddress = stringArr[0]
		}
	}
	return ipAddress
}
