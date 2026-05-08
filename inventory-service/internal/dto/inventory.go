package dto

import commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"

type Actor struct {
	UserID int64
	Role   commonv1.UserRole
}

type CreatePvzInput struct {
	Actor           Actor
	City            string
	Address         string
	CompanyName     string
	ContactPhone    string
	ContactTelegram *string
	Description     *string
	Lat             *float64
	Lng             *float64
}

type GetPvzInput struct {
	Actor Actor
	ID    int64
}

type ListOwnerPvzInput struct {
	Actor     Actor
	PageSize  int32
	PageToken *string
}

type UpdatePvzInput struct {
	Actor           Actor
	ID              int64
	City            *string
	Address         *string
	CompanyName     *string
	ContactPhone    *string
	ContactTelegram *string
	Description     *string
	Lat             *float64
	Lng             *float64
}

type CheckPvzForVacancyInput struct {
	Actor Actor
	PvzID int64
}

type ListPendingPvzInput struct {
	Actor     Actor
	PageSize  int32
	PageToken *string
}

type ModeratePvzInput struct {
	Actor   Actor
	PvzID   int64
	Reason  *string
	Comment *string
}

type Page struct {
	Limit  uint64
	Offset uint64
}
