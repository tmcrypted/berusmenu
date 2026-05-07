package service

import (
	"context"
	"errors"
	"testing"

	"github.com/tmcrypted/berusmenu/notification-service/internal/dto"
	"go.uber.org/mock/gomock"
)

func TestHandleEmailVerificationRequested(t *testing.T) {
	sendErr := errors.New("smtp unavailable")

	tests := []struct {
		name         string
		notification dto.EmailVerificationRequested
		setup        func(*MockSender)
		wantErr      error
	}{
		{
			name: "valid email verification notification",
			notification: dto.EmailVerificationRequested{
				UserID: 7,
				Email:  "worker@example.com",
				Code:   "123456",
			},
			setup: func(sender *MockSender) {
				sender.EXPECT().SendEmailVerification(gomock.Any(), dto.EmailVerificationRequested{
					UserID: 7,
					Email:  "worker@example.com",
					Code:   "123456",
				}).Return(nil)
			},
		},
		{
			name: "missing fields",
			notification: dto.EmailVerificationRequested{
				UserID: 7,
				Email:  "",
				Code:   "123456",
			},
			setup:   func(sender *MockSender) {},
			wantErr: ErrInvalidNotification,
		},
		{
			name: "sender error propagation",
			notification: dto.EmailVerificationRequested{
				UserID: 7,
				Email:  "worker@example.com",
				Code:   "123456",
			},
			setup: func(sender *MockSender) {
				sender.EXPECT().SendEmailVerification(gomock.Any(), gomock.Any()).Return(sendErr)
			},
			wantErr: sendErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			sender := NewMockSender(ctrl)
			tt.setup(sender)

			svc := New(sender)
			err := svc.HandleEmailVerificationRequested(context.Background(), tt.notification)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestHandleResponseCreated(t *testing.T) {
	sendErr := errors.New("smtp unavailable")

	tests := []struct {
		name         string
		notification dto.ResponseCreated
		setup        func(*MockSender)
		wantErr      error
	}{
		{
			name: "valid response notification",
			notification: dto.ResponseCreated{
				ResponseID:   11,
				VacancyID:    22,
				OwnerUserID:  33,
				OwnerEmail:   "owner@example.com",
				WorkerUserID: 44,
				VacancyTitle: "Night shift",
				ShiftDate:    "2026-05-07",
				PayAmount:    3000,
			},
			setup: func(sender *MockSender) {
				sender.EXPECT().SendResponseCreated(gomock.Any(), gomock.Any()).Return(nil)
			},
		},
		{
			name: "missing owner email",
			notification: dto.ResponseCreated{
				ResponseID:   11,
				VacancyID:    22,
				OwnerUserID:  33,
				WorkerUserID: 44,
			},
			setup:   func(sender *MockSender) {},
			wantErr: ErrInvalidNotification,
		},
		{
			name: "sender error propagation",
			notification: dto.ResponseCreated{
				ResponseID:   11,
				VacancyID:    22,
				OwnerUserID:  33,
				OwnerEmail:   "owner@example.com",
				WorkerUserID: 44,
			},
			setup: func(sender *MockSender) {
				sender.EXPECT().SendResponseCreated(gomock.Any(), gomock.Any()).Return(sendErr)
			},
			wantErr: sendErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			sender := NewMockSender(ctrl)
			tt.setup(sender)

			svc := New(sender)
			err := svc.HandleResponseCreated(context.Background(), tt.notification)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
