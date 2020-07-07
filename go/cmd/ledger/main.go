package main

import (
	"net/http"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/config"
	"go.uber.org/zap"
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

	r := gin.New()
	r.Use(ginzap.Ginzap(logger, time.RFC3339, true))
	r.Use(ginzap.RecoveryWithZap(logger, true))

	r.GET("/", func(c *gin.Context) {
		logger.Info("new request", zap.String("clientIp", c.ClientIP()))
		c.JSON(http.StatusOK, gin.H{"data": "hello world"})
	})

	logger.Info("listening", zap.String("listenAddr", listenAddr))

	r.Run(listenAddr)
}
