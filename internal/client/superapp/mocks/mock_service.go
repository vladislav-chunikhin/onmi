package mocks

import (
	"context"
	"fmt"
	"time"

	"onmi/internal/model/superapp"
)

type SuperApp struct {
}

func NewSuperApp() *SuperApp {
	return &SuperApp{}
}

func (sa *SuperApp) GetLimits() (n uint64, p time.Duration) {
	return 10, 2 * time.Second
}

func (sa *SuperApp) Process(_ context.Context, _ superapp.Batch) error {
	<-time.After(100 * time.Millisecond) // some work
	fmt.Println("processed the batch")
	return nil
}
