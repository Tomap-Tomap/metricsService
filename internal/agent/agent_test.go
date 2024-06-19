package agent

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/DarkOmap/metricsService/internal/logger"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

type ClientMockedObject struct {
	mock.Mock
}

func (c *ClientMockedObject) SendBatch(context.Context, map[string]float64) error {
	args := c.Called()

	return args.Error(0)
}

func (c *ClientMockedObject) SendCounter(context.Context, string, int64) error {
	args := c.Called()

	return args.Error(0)
}

type MSModckedObkect struct {
	mock.Mock
}

func (ms *MSModckedObkect) ReadMemStats() error {
	args := ms.Called()

	return args.Error(0)
}

func (ms *MSModckedObkect) GetMap() map[string]float64 {
	args := ms.Called()

	return args.Get(0).(map[string]float64)
}

func TestNewAgent(t *testing.T) {
	c := new(ClientMockedObject)
	ms := new(MSModckedObkect)

	t.Run("positive test", func(t *testing.T) {
		gotA := NewAgent(c, ms, 10, 10)

		require.Equal(t, &Agent{
			reportInterval: 10,
			pollInterval:   10,
			client:         c,
			ms:             ms,
		}, gotA)
	})
}

func TestAgent_Run(t *testing.T) {
	t.Run("positive test", func(t *testing.T) {
		t.Parallel()
		c := new(ClientMockedObject)
		c.On("SendBatch").Return(nil)
		c.On("SendCounter").Return(nil)
		ms := new(MSModckedObkect)
		ms.On("ReadMemStats").Return(nil)
		ms.On("GetMap").Return(map[string]float64{"test": 1})

		a := NewAgent(c, ms, 1, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := a.Run(ctx)

		require.NoError(t, err)

		ms.AssertExpectations(t)
	})

	t.Run("negative test", func(t *testing.T) {
		t.Parallel()
		c := new(ClientMockedObject)
		c.On("SendBatch").Return(nil)
		c.On("SendCounter").Return(nil)
		ms := new(MSModckedObkect)
		ms.On("ReadMemStats").Return(fmt.Errorf("test error"))
		ms.On("GetMap").Return(map[string]float64{"test": 1})

		a := NewAgent(c, ms, 1, 1)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		err := a.Run(ctx)

		require.Error(t, err)

		ms.AssertExpectations(t)
	})
}

type testingSink struct {
	*bytes.Buffer
}

func (s *testingSink) Close() error { return nil }
func (s *testingSink) Sync() error  { return nil }

func TestAgent_sendBatch(t *testing.T) {
	sink := &testingSink{new((bytes.Buffer))}
	zap.RegisterSink("testingBatch", func(u *url.URL) (zap.Sink, error) { return sink, nil })
	logger.Initialize("INFO", "testingBatch://")

	t.Run("positive tets", func(t *testing.T) {
		c := new(ClientMockedObject)
		c.On("SendBatch").Return(nil)
		ms := new(MSModckedObkect)
		ms.On("GetMap").Return(map[string]float64{"test": 1})

		a := NewAgent(c, ms, 1, 1)

		a.sendMemStats(context.Background())

		logs := sink.String()

		require.Empty(t, logs)
	})

	t.Run("negative tets", func(t *testing.T) {
		c := new(ClientMockedObject)
		c.On("SendBatch").Return(fmt.Errorf("test error"))
		ms := new(MSModckedObkect)
		ms.On("GetMap").Return(map[string]float64{"test": 1})

		a := NewAgent(c, ms, 1, 1)

		a.sendMemStats(context.Background())

		logs := sink.String()

		require.NotEmpty(t, logs)
	})
}

func TestAgent_sendCounter(t *testing.T) {
	sink := &testingSink{new((bytes.Buffer))}
	zap.RegisterSink("testingCounter", func(u *url.URL) (zap.Sink, error) { return sink, nil })
	logger.Initialize("INFO", "testingCounter://")

	t.Run("positive tets", func(t *testing.T) {
		c := new(ClientMockedObject)
		c.On("SendCounter").Return(nil)
		ms := new(MSModckedObkect)

		a := NewAgent(c, ms, 1, 1)

		a.sendPollCount(context.Background())

		logs := sink.String()

		require.Empty(t, logs)
	})

	t.Run("negative tets", func(t *testing.T) {
		c := new(ClientMockedObject)
		c.On("SendCounter").Return(fmt.Errorf("test error"))
		ms := new(MSModckedObkect)

		a := NewAgent(c, ms, 1, 1)

		a.sendPollCount(context.Background())

		logs := sink.String()

		require.NotEmpty(t, logs)
	})
}
