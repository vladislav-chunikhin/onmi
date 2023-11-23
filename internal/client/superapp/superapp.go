package superapp

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
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
	once      sync.Once
}

func NewClient(cfg *config.ClientConfig, logger logger.Logger, transport Transport, amountOfBatches int) (*Client, error) {
	if cfg == nil {
		return nil, errNilConfig
	}

	if logger == nil || reflect.ValueOf(logger).IsNil() {
		return nil, errNilLogger
	}

	if transport == nil || reflect.ValueOf(transport).IsNil() {
		return nil, errNilTransport
	}

	client := &Client{
		transport: transport,
		cfg:       cfg,
		logger:    logger,
		batchCh:   make(chan superapp.Batch, amountOfBatches),
	}
	client.setLimits()

	return client, nil
}

func (c *Client) Start(ctx context.Context) {
	ticker := time.NewTicker(c.p)
	defer func() {
		ticker.Stop()
		c.CloseBatchCh()
	}()

	for {
		select {
		case <-ctx.Done():
			c.logger.Infof("processing interrupted by done signal")
			return
		case <-ticker.C:
			batch := c.dequeueBatch()
			if batch == nil || len(*batch) == 0 {
				c.logger.Infof("processing completed")
				return
			}

			/*	There is an assumption that the batch processing time can be longer than the ticker period.
				Thus, it makes sense to improve this code section, for example, by running the processing as a separate goroutine*/
			err := c.processBatch(ctx, batch)
			if err != nil && errors.Is(err, superapp.ErrBlocked) {
				c.setLimits() // if an error is a blocked error, then we will try to update the limits
				ticker.Stop()
				ticker = time.NewTicker(c.p)
			}
		}
	}
}

func (c *Client) CloseBatchCh() {
	c.once.Do(func() {
		close(c.batchCh)
		c.logger.Debugf("batch channel is closed")
	})
}

// Close closes connection to super app
func (c *Client) Close() {
	c.logger.Debugf("connection to super app closed")
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
	maxItems := int(c.n)

	// Extract a specific number of elements
	for len(batch) < maxItems {
		select {
		case itemBatch, ok := <-c.batchCh:
			if !ok && len(batch) != 0 {
				return &batch
			}

			if !ok {
				return nil
			}

			diff := maxItems - len(batch)
			if len(itemBatch) > diff {
				c.batchCh <- itemBatch[diff:]
				itemBatch = itemBatch[:diff]
			}

			batch = append(batch, itemBatch...)
			if len(batch) == maxItems {
				return &batch
			}
		default:
			if len(batch) != 0 {
				return &batch
			}
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
