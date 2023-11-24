package main

import (
	"context"
	"log"
	"strconv"

	"github.com/vardius/shutdown"

	"onmi/internal/client/superapp"
	"onmi/internal/client/superapp/mocks"
	"onmi/internal/config"
	model "onmi/internal/model/superapp"
	configPkg "onmi/pkg/config"
	"onmi/pkg/logger"
)

const (
	// It can be changed for testing.
	// If the value is greater than or equal to 3, different values will be used instead of the first and the second batches.
	// See: numOfItems1 and numOfItems2
	amountOfBatches      = 3
	defaultAmountOfItems = 8 // can be changed for testing

	defaultConfigFilePath = "./config/default.yaml"
)

var (
	numOfItems1 = defaultAmountOfItems + 5 // can be changed for testing
	numOfItems2 = 1                        // can be changed for testing
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	// init: config
	cfg := &config.Config{}
	if err := configPkg.LoadConfig(cfg, defaultConfigFilePath); err != nil {
		log.Fatalf("init config: %v", err)
	}

	// init: logger
	cLog, err := logger.New(logger.DebugLevel)
	if err != nil {
		log.Fatalf("init log: %v", err)
	}

	// init: external service mock
	superApp := mocks.NewSuperApp(cLog)

	// init: external service client
	client, err := superapp.NewClient(&cfg.ClientsConfig.SupperApp, cLog, superApp, amountOfBatches)
	if err != nil {
		cLog.Fatalf("init client: %v", err)
	}

	// put batches on a buffered channel
	batches := getBatches()
	for _, batch := range batches {
		client.Enqueue(batch)
	}

	// start processing by the external service
	go client.Start(ctx)

	stop := func() {
		cLog.Debugf("graceful shutdown starting...")
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

	items1 := make([]model.Item, 0, numOfItems1)
	for i := 0; i < numOfItems1; i++ {
		items1 = append(items1, model.Item{ID: strconv.Itoa(i)})
	}

	items2 := make([]model.Item, 0, numOfItems2)
	for i := 0; i < numOfItems2; i++ {
		items2 = append(items2, model.Item{ID: strconv.Itoa(i)})
	}

	hasDifferentValues := cap(batches) >= 3

	for i := 0; i < amountOfBatches; i++ {
		if hasDifferentValues {
			switch i {
			case 0:
				batches = append(batches, items1)
			case 1:
				batches = append(batches, items2)
			default:
				batches = append(batches, defaultItems)
			}
			continue
		}
		batches = append(batches, defaultItems)
	}

	return batches
}
