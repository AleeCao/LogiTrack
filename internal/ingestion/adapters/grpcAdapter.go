package adapters

import (
	"io"

	v1 "github.com/AleeCao/LogiTrack/gen/go/tracking/v1"
	"github.com/AleeCao/LogiTrack/internal/ingestion/domain"
	"github.com/AleeCao/LogiTrack/internal/ingestion/ports"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type GrpcAdapter struct {
	v1.UnimplementedTrackingServer
	locPrcs ports.LocationProcessor
}

func NewGrpcAdapter(locPrc ports.LocationProcessor) *GrpcAdapter {
	return &GrpcAdapter{locPrcs: locPrc}
}

func (s *GrpcAdapter) GetLocation(stream v1.Tracking_GetLocationServer) error {
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Errorf(codes.Unknown, "failed to receive: %v", err)
		}

		location := domain.Location{
			TruckID:     msg.TruckId,
			Longitude:   msg.Longitude,
			Latitude:    msg.Latitude,
			UpdatedAt:   msg.Timestamp.AsTime(),
			TruckStatus: msg.Status.String(),
		}
		statusOpt, err := s.locPrcs.ProcessLocation(stream.Context(), &location)
		if err != nil {
			return status.Errorf(codes.Unknown, "failed to publish location on producer: %v", err)
		}
		switch statusOpt {
		case domain.DecisionContinue:
			err := stream.Send(&v1.LocationResponse{Status: v1.DeliveryStatus_CONTINUE})
			if err != nil {
				return status.Errorf(codes.Unavailable, "failed to send response: %v", err)
			}

		case domain.DecisionAbort:
			err := stream.Send(&v1.LocationResponse{Status: v1.DeliveryStatus_ABORT})
			if err != nil {
				return status.Errorf(codes.Unavailable, "failed to send response: %v", err)
			}

		case domain.DecisionComeBack:
			err := stream.Send(&v1.LocationResponse{Status: v1.DeliveryStatus_COME_BACK})
			if err != nil {
				return status.Errorf(codes.Unavailable, "failed to send response: %v", err)
			}
		}
	}
}
