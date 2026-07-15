package delivery

import (
	"context"
	proto "github.com/aidostt/protos/gen/go/reservista/qr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"qrcode-generation-service/pkg/logger"
)

func (h *Handler) Generate(ctx context.Context, input *proto.GenerateRequest) (*proto.GenerateResponse, error) {
	if input.Content == "" {
		return nil, status.Error(codes.InvalidArgument, "content is required")
	}
	qr, err := h.services.QrCode.GenerateQR(input.GetContent())
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to generate QR")
	}
	return &proto.GenerateResponse{QR: qr}, nil
}

func (h *Handler) Scan(ctx context.Context, input *proto.ScanRequest) (*proto.ScanResponse, error) {
	if input.UserID == "" {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if input.ReservationID == "" {
		return nil, status.Error(codes.InvalidArgument, "reservation id is required")
	}
	user, restaurant, reservation, err := h.services.QrCode.ScanQR(ctx, input.GetReservationID())
	if err != nil {
		logger.Error(err)
		return nil, status.Error(codes.Internal, "failed to scan QR")
	}
	return &proto.ScanResponse{
		UserName:          user.Name,
		UserSurname:       user.Surname,
		UserPhone:         user.Phone,
		UserEmail:         user.Email,
		RestaurantName:    restaurant.Name,
		RestaurantAddress: restaurant.Address,
		RestaurantContact: restaurant.Contact,
		TableID:           reservation.Table,
		ReservationTime:   reservation.ReservationTime,
	}, nil
}
