package grpc_handler

import (
	"context"
	"errors"

	"github.com/Arzeeq/pvz-api/internal/dto"
	pb "github.com/Arzeeq/pvz-api/internal/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type PVZServicer interface {
	GetPVZs(ctx context.Context) []dto.PVZ
}

type PVZHandler struct {
	service PVZServicer
}

func NewPVZHandler(service PVZServicer) (*PVZHandler, error) {
	if service == nil {
		return nil, errors.New("nil value in constructor")
	}

	return &PVZHandler{service: service}, nil
}

func (h *PVZHandler) GetPVZList(ctx context.Context, req *pb.GetPVZListRequest) (*pb.GetPVZListResponse, error) {
	pvzDTOs := h.service.GetPVZs(ctx)

	pvzProtos := make([]*pb.PVZ, 0, len(pvzDTOs))
	for _, p := range pvzDTOs {
		pvzProto, err := convertDTOToProto(p)
		if err != nil {
			return nil, errors.New("failed to convert from dto")
		}
		pvzProtos = append(pvzProtos, pvzProto)
	}

	return &pb.GetPVZListResponse{
		Pvzs: pvzProtos,
	}, nil
}

func convertDTOToProto(p dto.PVZ) (*pb.PVZ, error) {
	var id string
	if p.Id != nil {
		id = p.Id.String()
	}

	var regDate *timestamppb.Timestamp
	if p.RegistrationDate != nil {
		regDate = timestamppb.New(*p.RegistrationDate)
	}

	return &pb.PVZ{
		Id:               id,
		RegistrationDate: regDate,
		City:             string(p.City),
	}, nil
}
