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
)

// Transport defines external service that can process batches of items.
//
//go:generate mockery --name=Transport --with-expecter --case=underscore
type Transport interface {
	GetLimits() (n uint64, p time.Duration)
	Process(ctx context.Context, batch superapp.Batch) error
}

type Client struct {
	transport           Transport
	cfg                 *config.ClientConfig
	n                   uint64        // a certain number of elements that can be processed
	p                   time.Duration // the specified time interval from external service
	batchCh             chan superapp.Batch
	newTickerIntervalCh chan struct{}
	logger              logger.Logger
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
		transport:           transport,
		cfg:                 cfg,
		logger:              logger,
		batchCh:             make(chan superapp.Batch),
		newTickerIntervalCh: make(chan struct{}),
	}
	client.setLimits()

	return client, nil
}

func (c *Client) Start(ctx context.Context) {
	ticker := time.NewTicker(c.p)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if batch := c.dequeueBatch(); len(batch) > 0 {
					go func() {
						err := c.processBatch(ctx, batch)
						if err != nil && errors.Is(err, superapp.ErrBlocked) {
							c.setLimits()
							c.newTickerIntervalCh <- struct{}{}
						}
					}()
				}
			case <-c.newTickerIntervalCh:
				ticker.Stop()
				ticker = time.NewTicker(c.p)
			}
		}
	}()
}

func (c *Client) Close() {
	close(c.batchCh)
}

func (c *Client) setLimits() {
	n, p := c.transport.GetLimits()
	c.n = n
	c.p = p + c.cfg.Timeout // added time to account for network delays
}

func (c *Client) Enqueue(ub superapp.Batch) {
	c.batchCh <- ub
}

func (c *Client) dequeueBatch() superapp.Batch {
	batch := make(superapp.Batch, 0, c.n)

	for itemBatch := range c.batchCh {
		if len(batch)+len(itemBatch) <= int(c.n) {
			batch = append(batch, itemBatch...)
		} else {
			c.batchCh <- itemBatch
			break
		}
	}

	return batch
}

func (c *Client) processBatch(ctx context.Context, batch superapp.Batch) error {
	if err := c.transport.Process(ctx, batch); err != nil {
		c.logger.Errorf("failed to process the batch: %v", err)
		return fmt.Errorf("failed to process the batch: %w", err)
	}

	return nil
}
