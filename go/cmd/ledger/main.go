package main

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/app/ledger"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/config"
	_ "github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/swaggerui/statik"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/zaplog"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var listenAddr string

func init() {
	config.Prefix("OCVAB_LEDGER_")
	config.StringVar(&listenAddr, "listen", ":8080", "The address to listen on")
}

func main() {
	config.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	r := chi.NewRouter()
	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var fields []zapcore.Field
			for k, v := range r.Header {
				for i := range v {
					fields = append(fields, zap.String(fmt.Sprintf("%s[%d]", k, i), v[i]))
				}
			}
			logger.Info("header", fields...)
			next.ServeHTTP(w, r)
		})
	})
	r.Use(zaplog.ZapLog(logger))
	r.Use(middleware.Recoverer)

	r.Mount("/", ledger.Handler())

	logger.Info("listening", zap.String("listenAddr", listenAddr))

	http.ListenAndServe(listenAddr, r)
}
