package agent

import (
	"testing"

	"github.com/DarkOmap/metricsService/internal/client"
	"github.com/stretchr/testify/assert"
)

func TestNewAgent(t *testing.T) {
	testClient := client.NewClient("test")

	type args struct {
		client         *client.Client
		reportInterval uint
		pollInterval   uint
	}
	tests := []struct {
		name  string
		args  args
		wantA *Agent
	}{
		{
			name:  "positive test",
			args:  args{testClient, 10, 10},
			wantA: &Agent{reportInterval: 10, pollInterval: 10, client: testClient},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotA := NewAgent(tt.args.client, tt.args.reportInterval, tt.args.pollInterval)
			tt.wantA.ms = gotA.ms
			assert.Equal(t, tt.wantA, gotA)
		})
	}
}
