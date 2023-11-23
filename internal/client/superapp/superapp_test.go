package superapp

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"onmi/internal/client/superapp/mocks"
	"onmi/internal/config"
	"onmi/internal/model/superapp"
	"onmi/pkg/logger"
)

func TestTestSuperAppClient(t *testing.T) {
	suite.Run(t, new(TestSuperAppClientSuite))
}

type TestSuperAppClientSuite struct {
	suite.Suite

	cfg       *config.ClientConfig
	transport *mocks.Transport
	logger    logger.Logger
}

func (s *TestSuperAppClientSuite) SetupSuite() {
	s.cfg = &config.ClientConfig{
		Host:    "localhost",
		Port:    "80",
		Timeout: 5 * time.Second,
	}
	s.transport = mocks.NewTransport(s.T())

	log, err := logger.New(logger.DebugLevel)
	s.NoError(err)
	s.logger = log
}

func (s *TestSuperAppClientSuite) TearDownTest() {
	s.transport.ExpectedCalls = nil
}

func (s *TestSuperAppClientSuite) TestNewClient() {
	type args struct {
		cfg       *config.ClientConfig
		logger    logger.Logger
		transport *mocks.Transport
	}

	testCases := []struct {
		name        string
		applyMocks  func()
		args        *args
		want        *Client
		wantErr     bool
		expectedErr error
	}{
		{
			name: "nil transport",
			args: &args{
				cfg:       s.cfg,
				logger:    s.logger,
				transport: nil,
			},
			wantErr:     true,
			expectedErr: errNilTransport,
		},
		{
			name: "nil cfg",
			args: &args{
				cfg:    nil,
				logger: s.logger,
			},
			wantErr:     true,
			expectedErr: errNilConfig,
		},
		{
			name: "nil logger",
			args: &args{
				cfg:    s.cfg,
				logger: nil,
			},
			wantErr:     true,
			expectedErr: errNilLogger,
		},
		{
			name: "ok",
			applyMocks: func() {
				s.transport.EXPECT().
					GetLimits().
					Return(10, 5*time.Second).
					Once()
			},
			args: &args{
				cfg:       s.cfg,
				logger:    s.logger,
				transport: s.transport,
			},
			want: &Client{
				transport: s.transport,
				cfg:       s.cfg,
				n:         10,
				p:         5 * time.Second,
				batchCh:   make(chan superapp.Batch),
				logger:    s.logger,
			},
		},
	}

	for _, tt := range testCases {
		s.Run(tt.name, func() {
			if tt.applyMocks != nil {
				tt.applyMocks()
			}
			client, err := NewClient(tt.args.cfg, tt.args.logger, tt.args.transport, 5)
			if s.Assert().ErrorIs(err, tt.expectedErr) {
				ExpectError(&s.Suite, tt.expectedErr, err)
				return
			}

			s.Condition(func() (success bool) {
				success = client.cfg != tt.want.cfg ||
					client.transport != tt.want.transport ||
					client.batchCh == nil ||
					client.logger != tt.want.logger ||
					client.p != tt.want.p ||
					client.n != tt.want.n
				return !success
			}, "NewClient() got = %v, want %v", client, tt.want)
		})
	}
}

func (s *TestSuperAppClientSuite) TestClient_Start_OK() {
	ctx := context.TODO()
	amountOfBatches := 5
	amountOfItems := 3

	s.transport.EXPECT().
		GetLimits().
		Return(10, 2*time.Second).
		Once()

	s.transport.EXPECT().
		Process(mock.Anything, mock.Anything).
		Return(nil).
		Times(2)

	client, err := NewClient(s.cfg, s.logger, s.transport, amountOfBatches)
	s.NoError(err)

	batches := getBatches(amountOfBatches, amountOfItems)

	for _, batch := range batches {
		client.Enqueue(batch)
	}
	client.CloseBatchCh()

	client.Start(ctx)
}

func (s *TestSuperAppClientSuite) TestClient_Start_BlockError() {
	ctx := context.TODO()
	amountOfBatches := 3
	amountOfItems := 3

	s.transport.EXPECT().
		GetLimits().
		Return(10, 2*time.Second).
		Times(2)

	s.transport.EXPECT().
		Process(mock.Anything, mock.Anything).
		Return(superapp.ErrBlocked)

	client, err := NewClient(s.cfg, s.logger, s.transport, amountOfBatches)
	s.NoError(err)

	batches := getBatches(amountOfBatches, amountOfItems)

	for _, batch := range batches {
		client.Enqueue(batch)
	}
	client.CloseBatchCh()

	client.Start(ctx)
}

func (s *TestSuperAppClientSuite) TestClient_processBatch_Nil_Batch() {
	ctx := context.TODO()

	s.transport.EXPECT().
		GetLimits().
		Return(10, 2*time.Second).
		Once()

	client, err := NewClient(s.cfg, s.logger, s.transport, 5)
	s.NoError(err)

	err = client.processBatch(ctx, nil)
	s.Error(err)
	s.ErrorIs(err, errNilBatch)
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

func ExpectError(s *suite.Suite, expected error, actual error) {
	if expected == nil && actual == nil {
		return
	}
	if expected == nil {
		s.Equal(nil, actual)
		return
	}
	if actual == nil {
		s.Equal(expected, nil)
		return
	}

	s.Equal(expected.Error(), actual.Error())
}
