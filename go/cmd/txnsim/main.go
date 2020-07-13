package main

import (
	"context"
	"time"

	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/config"
	"github.com/opencredo/venafi-cloud-ab-poc/go/internal/pkg/ledgerserver"
	"go.uber.org/zap"
)

var serverAddr string

func init() {
	config.Prefix("OCVAB_TXNSIM_")
	config.StringVar(&serverAddr, "server", "localhost:8080", "The address to connect to")
}

func main() {
	config.Parse()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	api, err := ledgerserver.NewClientWithResponses(serverAddr)
	if err != nil {
		logger.Error("unable to create client", zap.Error(err), zap.String("serverAddr", serverAddr))
		return
	}

	ctx := context.Background()
	for {
		resp, err := api.PostTransactionsWithResponse(ctx, ledgerserver.PostTransactionsJSONRequestBody{
			Description: "random stuff",
			Amount:      123.45,
			FromAcct:    123456,
			ToAcct:      123457,
			Type:        "shopping",
		})
		if err != nil {
			logger.Error("unable to post transaction", zap.Error(err), zap.String("serverAddr", serverAddr))
			return
		}
		if resp.JSON400 != nil {
			logger.Warn("bad message response", zap.String("error", *resp.JSON400.Error))
		} else {
			location, err := resp.HTTPResponse.Location()
			if err != nil {
				logger.Error("missing location on response", zap.Error(err))
				return
			}

			logger.Info("success", zap.String("location", location.String()))
		}

		time.Sleep(1 * time.Minute)
	}
}
