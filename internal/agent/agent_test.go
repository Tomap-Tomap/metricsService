package agent

// type ClientMockedObject struct {
// 	mock.Mock
// }

// func (c *ClientMockedObject) SendBatch(ctx context.Context, batch map[string]float64) error {
// 	args := c.Called(batch)

// 	return args.Error(0)
// }

// func (c *ClientMockedObject) SendCounter(ctx context.Context, name string, delta int64) error {
// 	args := c.Called(delta)

// 	return args.Error(0)
// }

// func TestNewAgent(t *testing.T) {
// 	testClient := client.NewClient("test", "")

// 	type args struct {
// 		client         *client.Client
// 		reportInterval uint
// 		pollInterval   uint
// 		reportLimit    uint
// 	}
// 	tests := []struct {
// 		name  string
// 		args  args
// 		wantA *Agent
// 	}{
// 		{
// 			name:  "positive test",
// 			args:  args{testClient, 10, 10, 10},
// 			wantA: &Agent{reportInterval: 10, pollInterval: 10, rateLimit: 10, client: testClient},
// 		},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			gotA, err := NewAgent(tt.args.client, tt.args.reportInterval, tt.args.pollInterval, tt.args.reportLimit)
// 			require.NoError(t, err)
// 			gotA.ms = nil
// 			assert.Equal(t, tt.wantA, gotA)
// 		})
// 	}
// }

// type testingSink struct {
// 	*bytes.Buffer
// }

// func (s *testingSink) Close() error { return nil }
// func (s *testingSink) Sync() error  { return nil }

// func TestWorker(t *testing.T) {
// 	sink := &testingSink{new((bytes.Buffer))}
// 	zap.RegisterSink("testingWorker", func(u *url.URL) (zap.Sink, error) { return sink, nil })
// 	logger.Initialize("INFO", "testingWorker://")

// 	t.Run("no error test", func(t *testing.T) {
// 		jobs := make(chan func(context.Context) error, 1)
// 		jobs <- func(context.Context) error { return nil }
// 		close(jobs)
// 		worker(context.Background(), jobs)
// 		logs := sink.String()

// 		require.Empty(t, logs)
// 	})

// 	t.Run("error test", func(t *testing.T) {
// 		jobs := make(chan func(context.Context) error, 1)
// 		jobs <- func(context.Context) error { return fmt.Errorf("func error") }
// 		close(jobs)
// 		worker(context.Background(), jobs)

// 		logs := sink.String()

// 		require.Contains(t, logs, "func error")
// 	})
// }

// func TestAgent_sendBatch(t *testing.T) {
// 	agent := &Agent{}
// 	msNoError, err := memstats.NewMemStatsForServer()
// 	require.NoError(t, err)

// 	msError, err := memstats.NewMemStatsForServer()
// 	require.NoError(t, err)

// 	cm := new(ClientMockedObject)
// 	cm.On("SendBatch", msNoError.GetMap()).Return(nil)
// 	cm.On("SendBatch", msError.GetMap()).Return(fmt.Errorf("test error"))

// 	agent.client = cm
// 	t.Run("test no error", func(t *testing.T) {
// 		agent.ms = msNoError

// 		err := agent.sendBatch(context.Background())

// 		require.NoError(t, err)
// 	})

// 	t.Run("test error", func(t *testing.T) {
// 		agent.ms = msError

// 		err := agent.sendBatch(context.Background())

// 		require.Error(t, err)
// 	})

// 	cm.AssertExpectations(t)
// }

// func TestAgent_sendCounter(t *testing.T) {
// 	agent := &Agent{}

// 	deltaNoError := int64(0)
// 	deltaError := int64(1)

// 	cm := new(ClientMockedObject)
// 	cm.On("SendCounter", deltaNoError).Return(nil)
// 	cm.On("SendCounter", deltaError).Return(fmt.Errorf("test error"))

// 	agent.client = cm

// 	t.Run("test no error", func(t *testing.T) {
// 		agent.pollCount.Store(deltaNoError)
// 		err := agent.sendCounter(context.Background())

// 		require.NoError(t, err)
// 	})

// 	t.Run("test error", func(t *testing.T) {
// 		agent.pollCount.Store(deltaError)
// 		err := agent.sendCounter(context.Background())

// 		require.Error(t, err)
// 	})

// 	cm.AssertExpectations(t)
// }
