package mocks

import (
	"context"
	"fmt"
	"time"

	"onmi/internal/model/superapp"
	"onmi/pkg/logger"
)

const (
	nValue = 10
	pValue = 2 * time.Second
)

type SuperApp struct {
	log logger.Logger
}

func NewSuperApp(log logger.Logger) *SuperApp {
	return &SuperApp{log: log}
}

func (sa *SuperApp) GetLimits() (n uint64, p time.Duration) {
	return nValue, pValue
}

func (sa *SuperApp) Process(_ context.Context, batch superapp.Batch) error {
	if len(batch) > nValue {
		return fmt.Errorf("too many items")
	}

	<-time.After(100 * time.Millisecond) // some work
	sa.log.Debugf("processed the batch")
	return nil
}
