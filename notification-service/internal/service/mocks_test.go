package service

import (
	"context"
	"reflect"

	"github.com/tmcrypted/berusmenu/notification-service/internal/dto"
	"go.uber.org/mock/gomock"
)

type MockSender struct {
	ctrl     *gomock.Controller
	recorder *MockSenderMockRecorder
}

type MockSenderMockRecorder struct {
	mock *MockSender
}

func NewMockSender(ctrl *gomock.Controller) *MockSender {
	mock := &MockSender{ctrl: ctrl}
	mock.recorder = &MockSenderMockRecorder{mock}
	return mock
}

func (m *MockSender) EXPECT() *MockSenderMockRecorder {
	return m.recorder
}

func (m *MockSender) SendEmailVerification(ctx context.Context, notification dto.EmailVerificationRequested) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendEmailVerification", ctx, notification)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockSenderMockRecorder) SendEmailVerification(ctx, notification any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendEmailVerification", reflect.TypeOf((*MockSender)(nil).SendEmailVerification), ctx, notification)
}

func (m *MockSender) SendResponseCreated(ctx context.Context, notification dto.ResponseCreated) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendResponseCreated", ctx, notification)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockSenderMockRecorder) SendResponseCreated(ctx, notification any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendResponseCreated", reflect.TypeOf((*MockSender)(nil).SendResponseCreated), ctx, notification)
}
