package main

import (
	"net/http"
	"path"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/app/secretfixer"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/config"
	_ "github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/swaggerui/statik"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/zaplog"
	"go.uber.org/zap"
)

var listenAddr string
var certDir string

func init() {
	config.Prefix("OCVAB_SECRETFIXER_")
	config.StringVar(&listenAddr, "listen", ":8080", "The address to listen on")
	config.StringVar(&certDir, "certs", "", "Directory containing TLS certificates")
}

func main() {
	config.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(zaplog.ZapLog(logger))
	r.Use(middleware.Recoverer)

	r.Mount("/", secretfixer.Handler(logger))

	if len(certDir) > 0 {
		logger.Info("listening", zap.Bool("tls", true), zap.String("listenAddr", listenAddr))
		logger.Fatal("server exit", zap.Error(
			http.ListenAndServeTLS(listenAddr, path.Join(certDir, "tls.crt"), path.Join(certDir, "tls.key"), r)))
	} else {
		logger.Info("listening", zap.Bool("tls", false), zap.String("listenAddr", listenAddr))
		logger.Fatal("server exit", zap.Error(http.ListenAndServe(listenAddr, r)))
	}
}
