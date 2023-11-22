package superapp

import (
	"context"
	"errors"
	"fmt"
	"time"

	"onmi/internal/config"
	"onmi/internal/model/superapp"
	"onmi/pkg/logger"
)

var (
	errNilConfig    = fmt.Errorf("nil cfg")
	errNilLogger    = fmt.Errorf("nil logger")
	errNilTransport = fmt.Errorf("nil transport")
	errNilBatch     = fmt.Errorf("nil batch")
)

// Transport defines external service that can process batches of items.
//
//go:generate mockery --name=Transport --with-expecter --case=underscore
type Transport interface {
	GetLimits() (n uint64, p time.Duration)
	Process(ctx context.Context, batch superapp.Batch) error
}

type Client struct {
	transport Transport
	cfg       *config.ClientConfig
	n         uint64        // a certain number of elements that can be processed
	p         time.Duration // the specified time interval from external service
	batchCh   chan superapp.Batch
	logger    logger.Logger
}

func NewClient(cfg *config.ClientConfig, logger logger.Logger, transport Transport) (*Client, error) {
	if cfg == nil {
		return nil, errNilConfig
	}

	if logger == nil {
		return nil, errNilLogger
	}

	if transport == nil {
		return nil, errNilTransport
	}

	client := &Client{
		transport: transport,
		cfg:       cfg,
		logger:    logger,
		batchCh:   make(chan superapp.Batch),
	}
	client.setLimits()

	return client, nil
}

func (c *Client) Start(ctx context.Context) {
	ticker := time.NewTicker(c.p)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			batch := c.dequeueBatch()
			if batch == nil || len(*batch) == 0 {
				return
			}

			err := c.processBatch(ctx, batch)
			if err != nil && errors.Is(err, superapp.ErrBlocked) {
				c.setLimits() // if an error is a blocked error, then we will try to update the limits
				ticker.Stop()
				ticker = time.NewTicker(c.p)
			}
		}
	}
}

func (c *Client) Close() {
	close(c.batchCh)
}

func (c *Client) setLimits() {
	n, p := c.transport.GetLimits()
	c.n = n
	c.p = p + c.cfg.Delta // added time to account for network delays
}

func (c *Client) Enqueue(ub superapp.Batch) {
	c.batchCh <- ub
}

func (c *Client) dequeueBatch() *superapp.Batch {
	batch := make(superapp.Batch, 0, c.n)

	// Extract a specific number of elements
	for len(batch) < int(c.n) {
		select {
		case itemBatch, ok := <-c.batchCh:
			if !ok && len(batch) != 0 {
				return &batch
			}

			if !ok {
				fmt.Println("processing completed...")
				return nil
			}

			batch = append(batch, itemBatch...)
		default:
			return nil // No available items in the channel, finish extraction
		}
	}

	return &batch
}

func (c *Client) processBatch(ctx context.Context, batch *superapp.Batch) error {
	if batch == nil {
		return errNilBatch
	}

	requestCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer func() {
		if requestCtx.Err() == nil {
			cancel()
		}
	}()
	if err := c.transport.Process(requestCtx, *batch); err != nil {
		c.logger.Errorf("failed to process the batch: %v", err)
		return fmt.Errorf("failed to process the batch: %w", err)
	}

	return nil
}
