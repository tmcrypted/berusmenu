package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/tmcrypted/berusmenu/notification-service/internal/dto"
)

var ErrInvalidNotification = errors.New("invalid notification")

type Sender interface {
	SendEmailVerification(ctx context.Context, notification dto.EmailVerificationRequested) error
	SendResponseCreated(ctx context.Context, notification dto.ResponseCreated) error
}

type Service struct {
	sender Sender
}

func New(sender Sender) *Service {
	return &Service{sender: sender}
}

func (s *Service) HandleEmailVerificationRequested(ctx context.Context, notification dto.EmailVerificationRequested) error {
	if notification.UserID <= 0 || strings.TrimSpace(notification.Email) == "" || strings.TrimSpace(notification.Code) == "" {
		return ErrInvalidNotification
	}
	return s.sender.SendEmailVerification(ctx, notification)
}

func (s *Service) HandleResponseCreated(ctx context.Context, notification dto.ResponseCreated) error {
	if notification.ResponseID <= 0 || notification.VacancyID <= 0 || notification.OwnerUserID <= 0 ||
		strings.TrimSpace(notification.OwnerEmail) == "" || notification.WorkerUserID <= 0 {
		return ErrInvalidNotification
	}
	return s.sender.SendResponseCreated(ctx, notification)
}

type LogSender struct{}

func NewLogSender() *LogSender {
	return &LogSender{}
}

func (s *LogSender) SendEmailVerification(ctx context.Context, notification dto.EmailVerificationRequested) error {
	_ = ctx
	log.Println("==================================================")
	log.Printf("[EMAIL STUB] To: %s", notification.Email)
	log.Printf("[EMAIL STUB] Verification Code: %s", notification.Code)
	log.Printf("[EMAIL STUB] User ID: %d", notification.UserID)
	log.Println("==================================================")
	return nil
}

func (s *LogSender) SendResponseCreated(ctx context.Context, notification dto.ResponseCreated) error {
	_ = ctx
	log.Println("==================================================")
	log.Printf("[EMAIL STUB] To: %s", notification.OwnerEmail)
	log.Printf("[EMAIL STUB] New response: response_id=%d vacancy_id=%d worker_id=%d", notification.ResponseID, notification.VacancyID, notification.WorkerUserID)
	log.Printf("[EMAIL STUB] Vacancy: %s shift=%s pay=%d", notification.VacancyTitle, notification.ShiftDate, notification.PayAmount)
	log.Println("==================================================")
	return nil
}

func FormatEmailVerificationStub(notification dto.EmailVerificationRequested) string {
	return fmt.Sprintf("[EMAIL STUB] To: %s Verification Code: %s User ID: %d", notification.Email, notification.Code, notification.UserID)
}

func FormatResponseCreatedStub(notification dto.ResponseCreated) string {
	return fmt.Sprintf("[EMAIL STUB] To: %s New response %d for vacancy %d", notification.OwnerEmail, notification.ResponseID, notification.VacancyID)
}
