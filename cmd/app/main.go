package main

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/vardius/shutdown"

	"onmi/internal/client/superapp"
	"onmi/internal/client/superapp/mocks"
	"onmi/internal/config"
	model "onmi/internal/model/superapp"
	"onmi/pkg/logger"
)

const (
	amountOfBatches = 3
	amountOfItems   = 10
)

func main() {
	superApp := mocks.NewSuperApp()
	ctx, cancel := context.WithCancel(context.Background())

	cfg := &config.Config{
		ClientsConfig: config.ClientsConfig{
			SupperApp: config.ClientConfig{
				Host:    "localhost",
				Port:    "8080",
				Timeout: 5 * time.Second,
				Delta:   time.Second,
			},
		},
	}

	customLog, err := logger.New(logger.DebugLevel)
	if err != nil {
		log.Fatalf("init log: %v", err)
	}

	batches := getBatches(amountOfBatches, amountOfItems)

	client, err := superapp.NewClient(&cfg.ClientsConfig.SupperApp, customLog, superApp, amountOfBatches)
	if err != nil {
		log.Fatalf("init client: %v", err)
	}

	for _, batch := range batches {
		client.Enqueue(batch)
	}
	client.CloseBatchCh()

	go client.Start(ctx)

	stop := func() {
		client.Close()
		cancel()
	}

	shutdown.GracefulStop(stop)
}

func getBatches(amountOfBatches, amountOfItems int) []model.Batch {
	batches := make([]model.Batch, 0, amountOfBatches)
	items := make([]model.Item, 0, amountOfItems)

	for i := 0; i < amountOfItems; i++ {
		items = append(items, model.Item{ID: strconv.Itoa(i)})
	}

	for i := 0; i < amountOfBatches; i++ {
		batches = append(batches, items)
	}

	return batches
}
