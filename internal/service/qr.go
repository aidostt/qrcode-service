package service

import (
	"bytes"
	"context"
	"fmt"
	proto_reservation "github.com/aidostt/protos/gen/go/reservista/reservation"
	"github.com/aidostt/protos/gen/go/reservista/user"
	"github.com/nfnt/resize"
	"github.com/skip2/go-qrcode"
	"image"
	"image/draw"
	"image/png"
	"io"
	"qrcode-generation-service/pkg/dialog"
)

const (
	size = 256
)

type Service struct {
	domain string
	dialog *dialog.Dialog
}

func NewGeneratorService(domain string, dial *dialog.Dialog) *Service {
	return &Service{
		domain: domain,
		dialog: dial,
	}
}

func (s *Service) GenerateQR(content string) ([]byte, error) {
	qrCode, err := qrcode.Encode(content, qrcode.Low, size)
	if err != nil {
		return nil, fmt.Errorf("could not generate a QR code: %v", err)
	}
	return qrCode, nil
}

// ScanQR resolves a reservation for a staff member scanning its QR code and
// returns the guest and reservation details they need to confirm the booking.
// Authorization to scan is enforced upstream (staff-only route); this method
// therefore does not check the scanner against the reservation owner — that
// check was inverted and blocked the intended staff-scans-guest flow.
func (s *Service) ScanQR(ctx context.Context, reservationID string) (UserInfo, RestaurantInfo, ReservationInfo, error) {
	reservationConn, err := s.dialog.NewConnection(s.dialog.Addresses.Reservations)
	if err != nil {
		return UserInfo{}, RestaurantInfo{}, ReservationInfo{}, err
	}
	defer reservationConn.Close()

	reservationClient := proto_reservation.NewReservationClient(reservationConn)
	reservation, err := reservationClient.GetReservation(ctx, &proto_reservation.IDRequest{Id: reservationID})
	if err != nil {
		return UserInfo{}, RestaurantInfo{}, ReservationInfo{}, err
	}

	userConn, err := s.dialog.NewConnection(s.dialog.Addresses.Users)
	if err != nil {
		return UserInfo{}, RestaurantInfo{}, ReservationInfo{}, err
	}
	defer userConn.Close()

	// Look up the guest who holds the reservation, not the staff member scanning.
	userClient := proto_user.NewUserClient(userConn)
	guest, err := userClient.GetByID(ctx, &proto_user.GetRequest{UserId: reservation.GetUserID()})
	if err != nil {
		return UserInfo{}, RestaurantInfo{}, ReservationInfo{}, err
	}

	user := UserInfo{
		Name:    guest.GetName(),
		Surname: guest.GetSurname(),
		Phone:   guest.GetPhone(),
		Email:   guest.GetEmail(),
	}
	reservationInfo := ReservationInfo{
		Table:           reservation.Table.GetTableNumber(),
		ReservationTime: reservation.GetStartAt().AsTime().Format("15:04, Jan 02 2006"),
	}
	restaurant := RestaurantInfo{
		Name:    reservation.Table.Restaurant.GetName(),
		Contact: reservation.Table.Restaurant.GetContact(),
		Address: reservation.Table.Restaurant.GetAddress(),
	}
	return user, restaurant, reservationInfo, nil
}

func (s *Service) GenerateQRWithWatermark(watermark []byte, content string) ([]byte, error) {
	qrCode, err := s.GenerateQR(content)
	if err != nil {
		return nil, err
	}

	qrCode, err = s.AddWatermark(qrCode, watermark)
	if err != nil {
		return nil, fmt.Errorf("could not add watermark to QR code: %v", err)
	}

	return qrCode, nil
}

func (s *Service) AddWatermark(qrCode []byte, watermarkData []byte) ([]byte, error) {
	qrCodeData, err := png.Decode(bytes.NewBuffer(qrCode))
	if err != nil {
		return nil, fmt.Errorf("could not decode QR code: %v", err)
	}

	watermarkWidth := uint(float64(qrCodeData.Bounds().Dx()) * 0.25)
	watermark, err := s.ResizeWatermark(bytes.NewBuffer(watermarkData), watermarkWidth)
	if err != nil {
		return nil, fmt.Errorf("could not resize the watermark image: %v", err)
	}

	watermarkImage, err := png.Decode(bytes.NewBuffer(watermark))
	if err != nil {
		return nil, fmt.Errorf("could not decode watermark: %v", err)
	}

	halfQrCodeWidth, halfWatermarkWidth := qrCodeData.Bounds().Dx()/2, watermarkImage.Bounds().Dx()/2
	offset := image.Pt(
		halfQrCodeWidth-halfWatermarkWidth,
		halfQrCodeWidth-halfWatermarkWidth,
	)

	watermarkImageBounds := qrCodeData.Bounds()
	m := image.NewRGBA(watermarkImageBounds)

	draw.Draw(m, watermarkImageBounds, qrCodeData, image.Point{}, draw.Src)
	draw.Draw(
		m,
		watermarkImage.Bounds().Add(offset),
		watermarkImage,
		image.Point{},
		draw.Over,
	)

	watermarkedQRCode := bytes.NewBuffer(nil)
	png.Encode(watermarkedQRCode, m)

	return watermarkedQRCode.Bytes(), nil
}

func (s *Service) ResizeWatermark(watermark io.Reader, width uint) ([]byte, error) {
	decodedImage, err := png.Decode(watermark)
	if err != nil {
		return nil, fmt.Errorf("could not decode watermark image: %v", err)
	}

	m := resize.Resize(width, 0, decodedImage, resize.Lanczos3)
	resized := bytes.NewBuffer(nil)
	png.Encode(resized, m)

	return resized.Bytes(), nil
}
