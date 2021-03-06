package main

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/app/ledger"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/config"
	_ "github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/swaggerui/statik"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/zaplog"
	"go.uber.org/zap"
)

var listenAddr string
var pgConnection string

func init() {
	config.Prefix("OCVAB_LEDGER_")
	config.StringVar(&listenAddr, "listen", ":8080", "The address to listen on")
	config.StringVar(&pgConnection, "db", "", "Postgres connection string. If empty ledger will be in memory.")
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

	if pgConnection != "" {
		err := ledger.InitDB(logger, pgConnection)
		if err != nil {
			logger.Fatal("unable to initialise database", zap.Error(err), zap.String("pgConnection", pgConnection))
		}
	}

	r.Mount("/", ledger.Handler())

	logger.Info("listening", zap.String("listenAddr", listenAddr))

	logger.Fatal("server exit", zap.Error(http.ListenAndServe(listenAddr, r)))
}
