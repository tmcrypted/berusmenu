package service

import (
	"context"
	"reflect"
	"time"

	"github.com/tmcrypted/berusmenu/auth-service/internal/store"
	"github.com/tmcrypted/berusmenu/auth-service/internal/verification"
	"go.uber.org/mock/gomock"
)

type MockUserRepository struct {
	ctrl     *gomock.Controller
	recorder *MockUserRepositoryMockRecorder
}

type MockUserRepositoryMockRecorder struct {
	mock *MockUserRepository
}

func NewMockUserRepository(ctrl *gomock.Controller) *MockUserRepository {
	mock := &MockUserRepository{ctrl: ctrl}
	mock.recorder = &MockUserRepositoryMockRecorder{mock}
	return mock
}

func (m *MockUserRepository) EXPECT() *MockUserRepositoryMockRecorder {
	return m.recorder
}

func (m *MockUserRepository) CreateUserWithSession(ctx context.Context, user store.User, refreshToken string, refreshExpiresAt time.Time) (store.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserWithSession", ctx, user, refreshToken, refreshExpiresAt)
	ret0, _ := ret[0].(store.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) CreateUserWithSession(ctx, user, refreshToken, refreshExpiresAt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserWithSession", reflect.TypeOf((*MockUserRepository)(nil).CreateUserWithSession), ctx, user, refreshToken, refreshExpiresAt)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (store.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmail", ctx, email)
	ret0, _ := ret[0].(store.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetUserByEmail(ctx, email any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmail", reflect.TypeOf((*MockUserRepository)(nil).GetUserByEmail), ctx, email)
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id int64) (store.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByID", ctx, id)
	ret0, _ := ret[0].(store.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) GetUserByID(ctx, id any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByID", reflect.TypeOf((*MockUserRepository)(nil).GetUserByID), ctx, id)
}

func (m *MockUserRepository) CreateRefreshSession(ctx context.Context, userID int64, refreshToken string, expiresAt time.Time) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateRefreshSession", ctx, userID, refreshToken, expiresAt)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) CreateRefreshSession(ctx, userID, refreshToken, expiresAt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateRefreshSession", reflect.TypeOf((*MockUserRepository)(nil).CreateRefreshSession), ctx, userID, refreshToken, expiresAt)
}

func (m *MockUserRepository) RefreshSessionUser(ctx context.Context, refreshToken string) (store.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RefreshSessionUser", ctx, refreshToken)
	ret0, _ := ret[0].(store.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) RefreshSessionUser(ctx, refreshToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RefreshSessionUser", reflect.TypeOf((*MockUserRepository)(nil).RefreshSessionUser), ctx, refreshToken)
}

func (m *MockUserRepository) RotateRefreshSession(ctx context.Context, oldRefreshToken string, newRefreshToken string, newExpiresAt time.Time) (store.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RotateRefreshSession", ctx, oldRefreshToken, newRefreshToken, newExpiresAt)
	ret0, _ := ret[0].(store.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) RotateRefreshSession(ctx, oldRefreshToken, newRefreshToken, newExpiresAt any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RotateRefreshSession", reflect.TypeOf((*MockUserRepository)(nil).RotateRefreshSession), ctx, oldRefreshToken, newRefreshToken, newExpiresAt)
}

func (m *MockUserRepository) RevokeRefreshSession(ctx context.Context, refreshToken string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "RevokeRefreshSession", ctx, refreshToken)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockUserRepositoryMockRecorder) RevokeRefreshSession(ctx, refreshToken any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "RevokeRefreshSession", reflect.TypeOf((*MockUserRepository)(nil).RevokeRefreshSession), ctx, refreshToken)
}

func (m *MockUserRepository) MarkEmailVerified(ctx context.Context, userID int64) (store.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MarkEmailVerified", ctx, userID)
	ret0, _ := ret[0].(store.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) MarkEmailVerified(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MarkEmailVerified", reflect.TypeOf((*MockUserRepository)(nil).MarkEmailVerified), ctx, userID)
}

func (m *MockUserRepository) UpdateProfile(ctx context.Context, userID int64, fullName *string, phone *string, avatarURL *string) (store.User, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "UpdateProfile", ctx, userID, fullName, phone, avatarURL)
	ret0, _ := ret[0].(store.User)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockUserRepositoryMockRecorder) UpdateProfile(ctx, userID, fullName, phone, avatarURL any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "UpdateProfile", reflect.TypeOf((*MockUserRepository)(nil).UpdateProfile), ctx, userID, fullName, phone, avatarURL)
}

type MockVerificationStore struct {
	ctrl     *gomock.Controller
	recorder *MockVerificationStoreMockRecorder
}

type MockVerificationStoreMockRecorder struct {
	mock *MockVerificationStore
}

func NewMockVerificationStore(ctrl *gomock.Controller) *MockVerificationStore {
	mock := &MockVerificationStore{ctrl: ctrl}
	mock.recorder = &MockVerificationStoreMockRecorder{mock}
	return mock
}

func (m *MockVerificationStore) EXPECT() *MockVerificationStoreMockRecorder {
	return m.recorder
}

func (m *MockVerificationStore) CreateOrReuseCode(ctx context.Context, userID int64) (verification.CodeState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateOrReuseCode", ctx, userID)
	ret0, _ := ret[0].(verification.CodeState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockVerificationStoreMockRecorder) CreateOrReuseCode(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateOrReuseCode", reflect.TypeOf((*MockVerificationStore)(nil).CreateOrReuseCode), ctx, userID)
}

func (m *MockVerificationStore) ResendCode(ctx context.Context, userID int64) (verification.CodeState, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "ResendCode", ctx, userID)
	ret0, _ := ret[0].(verification.CodeState)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockVerificationStoreMockRecorder) ResendCode(ctx, userID any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "ResendCode", reflect.TypeOf((*MockVerificationStore)(nil).ResendCode), ctx, userID)
}

func (m *MockVerificationStore) VerifyCode(ctx context.Context, userID int64, submitted string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "VerifyCode", ctx, userID, submitted)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockVerificationStoreMockRecorder) VerifyCode(ctx, userID, submitted any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "VerifyCode", reflect.TypeOf((*MockVerificationStore)(nil).VerifyCode), ctx, userID, submitted)
}

type MockEventPublisher struct {
	ctrl     *gomock.Controller
	recorder *MockEventPublisherMockRecorder
}

type MockEventPublisherMockRecorder struct {
	mock *MockEventPublisher
}

func NewMockEventPublisher(ctrl *gomock.Controller) *MockEventPublisher {
	mock := &MockEventPublisher{ctrl: ctrl}
	mock.recorder = &MockEventPublisherMockRecorder{mock}
	return mock
}

func (m *MockEventPublisher) EXPECT() *MockEventPublisherMockRecorder {
	return m.recorder
}

func (m *MockEventPublisher) EmailVerificationRequested(ctx context.Context, userID int64, email string, code string) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "EmailVerificationRequested", ctx, userID, email, code)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockEventPublisherMockRecorder) EmailVerificationRequested(ctx, userID, email, code any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "EmailVerificationRequested", reflect.TypeOf((*MockEventPublisher)(nil).EmailVerificationRequested), ctx, userID, email, code)
}
