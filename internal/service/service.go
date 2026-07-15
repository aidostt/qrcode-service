package service

import (
	"context"
	"io"
	"qrcode-generation-service/pkg/dialog"
)

type RestaurantInfo struct {
	Name    string
	Address string
	Contact string
}

type ReservationInfo struct {
	Table           int32
	ReservationTime string
}

type UserInfo struct {
	Name    string
	Surname string
	Phone   string
	Email   string
}

type QRCode interface {
	GenerateQR(string) ([]byte, error)
	ScanQR(context.Context, string) (UserInfo, RestaurantInfo, ReservationInfo, error)
	GenerateQRWithWatermark([]byte, string) ([]byte, error)
	AddWatermark([]byte, []byte) ([]byte, error)
	ResizeWatermark(io.Reader, uint) ([]byte, error)
}

type Services struct {
	QrCode QRCode
}

type Dependencies struct {
	Environment string
	Domain      string
	Dialog      *dialog.Dialog
}

func NewServices(deps Dependencies) *Services {
	generatorService := NewGeneratorService(deps.Domain, deps.Dialog)
	return &Services{
		QrCode: generatorService,
	}
}
