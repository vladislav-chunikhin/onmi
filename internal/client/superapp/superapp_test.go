package superapp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"onmi/internal/client/superapp/mocks"
	"onmi/internal/config"
	"onmi/internal/model/superapp"
	"onmi/pkg/logger"
)

func TestNewClient_NilTransport(t *testing.T) {
	client, err := NewClient(&config.ClientConfig{
		Host: "localhost",
		Port: "80",
	}, mocks.NewMockLogger(), nil)

	require.Nil(t, client)
	require.Error(t, err)
	require.ErrorIs(t, err, errNilTransport)
}

func TestNewClient(t *testing.T) {
	type args struct {
		cfg       *config.ClientConfig
		logger    logger.Logger
		transport func(transport *mocks.Transport)
	}

	transport := mocks.NewTransport(t)

	okArgs := &args{
		cfg: &config.ClientConfig{
			Host: "localhost",
			Port: "80",
		},
		logger: mocks.NewMockLogger(),
		transport: func(transport *mocks.Transport) {
			transport.EXPECT().
				GetLimits().
				Return(10, 5*time.Second).
				Once()
		},
	}

	testCases := []struct {
		name        string
		args        *args
		want        *Client
		wantErr     bool
		expectedErr error
	}{
		{
			name: "nil cfg",
			args: &args{
				cfg:    nil,
				logger: mocks.NewMockLogger(),
			},
			wantErr:     true,
			expectedErr: errNilConfig,
		},
		{
			name: "nil logger",
			args: &args{
				cfg: &config.ClientConfig{
					Host: "localhost",
					Port: "80",
				},
				logger: nil,
			},
			wantErr:     true,
			expectedErr: errNilLogger,
		},
		{
			name: "ok",
			args: okArgs,
			want: &Client{
				transport: transport,
				cfg:       okArgs.cfg,
				n:         10,
				p:         5 * time.Second,
				batchCh:   make(chan superapp.Batch),
				logger:    okArgs.logger,
			},
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.transport != nil {
				tt.args.transport(transport)
			}
			client, err := NewClient(tt.args.cfg, tt.args.logger, transport)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil {
				require.ErrorIs(t, err, tt.expectedErr)
				return
			}

			if client.cfg != tt.want.cfg ||
				client.transport != tt.want.transport ||
				client.batchCh == nil ||
				client.logger != tt.want.logger ||
				client.p != tt.want.p ||
				client.n != tt.want.n {
				t.Errorf("NewClient() got = %v, want %v", client, tt.want)
			}
		})
	}
}

func TestClient_Start_OK(t *testing.T) {
	cfg := &config.ClientConfig{
		Host:    "localhost",
		Port:    "80",
		Timeout: 5 * time.Second,
	}
	transport := mocks.NewTransport(t)
	transport.EXPECT().
		GetLimits().
		Return(10, 10*time.Second).
		Once()

	ctx := context.TODO()
	transport.EXPECT().
		Process(mock.Anything, mock.Anything).
		Return(nil).
		Times(2)

	log := mocks.NewMockLogger()

	client, err := NewClient(cfg, log, transport)
	require.Nil(t, err)

	batches := getBatches(30, 10)

	go func() {
		for _, batch := range batches {
			client.Enqueue(batch)
		}
	}()

	client.Start(ctx)
}

func getBatches(amountOfBatches, amountOfItems int) []superapp.Batch {
	batches := make([]superapp.Batch, 0, amountOfBatches)
	items := make([]superapp.Item, 0, amountOfItems)

	for i := 0; i < amountOfItems; i++ {
		items = append(items, superapp.Item{})
	}

	for i := 0; i < amountOfBatches; i++ {
		batches = append(batches, items)
	}

	return batches
}
