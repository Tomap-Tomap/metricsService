package handlers

import (
	"context"

	"github.com/DarkOmap/metricsService/internal/models"
	"github.com/DarkOmap/metricsService/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	empty "github.com/golang/protobuf/ptypes/empty"
)

// MetricsServer structure with grpc methods
type MetricsServer struct {
	proto.UnimplementedMetricsServer

	r Repository
}

// NewMetricsServer create MetricsServer
func NewMetricsServer(r Repository) *MetricsServer {
	return &MetricsServer{r: r}
}

// Update sends a request to update metric
func (s *MetricsServer) Update(ctx context.Context, req *proto.UpdateRequest) (*proto.UpdateResponse, error) {
	var response proto.UpdateResponse

	m, err := models.NewMetricByProto(req.Metric)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	rM, err := s.r.UpdateByMetrics(ctx, *m)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	response.Metric = &proto.Metric{
		Id: rM.ID,
	}

	switch rM.MType {
	case "counter":
		response.Metric.Data = &proto.Metric_Delta{Delta: *rM.Delta}
		response.Metric.Type = proto.Types_COUNTER
	case "gauge":
		response.Metric.Data = &proto.Metric_Value{Value: *rM.Value}
		response.Metric.Type = proto.Types_GAUGE
	}

	return &response, nil
}

// Updates sends a request to update metrics
func (s *MetricsServer) Updates(ctx context.Context, req *proto.UpdatesRequest) (*empty.Empty, error) {
	ms, err := models.NewMetricsSliceByProto(req.Metrics)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	err = s.r.Updates(ctx, ms)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return nil, nil
}
