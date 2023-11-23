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
	amountOfBatches      = 3 // can be changed for testing
	defaultAmountOfItems = 8 // can be changed for testing

	hostValue    = "localhost"
	portValue    = "8080"
	timeoutValue = 5 * time.Second
	deltaValue   = 1 * time.Second
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// init: config
	cfg := &config.Config{
		ClientsConfig: config.ClientsConfig{
			SupperApp: config.ClientConfig{
				Host:    hostValue,
				Port:    portValue,
				Timeout: timeoutValue,
				Delta:   deltaValue,
			},
		},
	}

	// init: logger
	customLog, err := logger.New(logger.DebugLevel)
	if err != nil {
		log.Fatalf("init log: %v", err)
	}

	// init: external service mock
	superApp := mocks.NewSuperApp(customLog)

	// init: external service client
	client, err := superapp.NewClient(&cfg.ClientsConfig.SupperApp, customLog, superApp, amountOfBatches)
	if err != nil {
		customLog.Fatalf("init client: %v", err)
	}

	// put batches on a buffered channel
	batches := getBatches()
	for _, batch := range batches {
		client.Enqueue(batch)
	}

	// start processing by the external service
	go client.Start(ctx)

	stop := func() {
		customLog.Debugf("graceful shutdown starting...")
		client.CloseBatchCh()
		client.Close()
		cancel()
	}

	shutdown.GracefulStop(stop)
}

func getBatches() []model.Batch {
	batches := make([]model.Batch, 0, amountOfBatches)

	defaultItems := make([]model.Item, 0, defaultAmountOfItems)
	for i := 0; i < defaultAmountOfItems; i++ {
		defaultItems = append(defaultItems, model.Item{ID: strconv.Itoa(i)})
	}

	numOfItems1 := defaultAmountOfItems + 5 // can be changed for testing
	items1 := make([]model.Item, 0, numOfItems1)
	for i := 0; i < numOfItems1; i++ {
		items1 = append(items1, model.Item{ID: strconv.Itoa(i)})
	}

	numOfItems2 := 1 // can be changed for testing
	items2 := make([]model.Item, 0, numOfItems2)
	for i := 0; i < numOfItems2; i++ {
		items2 = append(items2, model.Item{ID: strconv.Itoa(i)})
	}

	for i := 0; i < amountOfBatches; i++ {
		switch i {
		case 0:
			batches = append(batches, items1)
		case 1:
			batches = append(batches, items2)
		default:
			batches = append(batches, defaultItems)
		}
	}

	return batches
}
