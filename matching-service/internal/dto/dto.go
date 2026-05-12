package dto

import (
	"time"

	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

type Actor struct {
	UserID int64
	Role   commonv1.UserRole
	Email  string
}

type Page struct {
	Limit  uint64
	Offset uint64
}

type UserSnapshot struct {
	UserID      int64
	Email       string
	DisplayName string
	Role        commonv1.UserRole
}

type PvzSnapshot struct {
	PvzID       int64
	OwnerID     int64
	CompanyName string
	Address     string
	City        string
}

type CreateVacancyInput struct {
	Actor        Actor
	PvzID        int64
	Title        *string
	Description  *string
	ShiftStartAt time.Time
	ShiftEndAt   time.Time
	PayAmount    int64
	SlotsTotal   int32
}

type ListVacanciesInput struct {
	Actor          Actor
	PvzID          *int64
	ShiftStartFrom *time.Time
	ShiftStartTo   *time.Time
	Status         commonv1.VacancyStatus
	PageSize       int32
	PageToken      *string
}

type IDInput struct {
	Actor Actor
	ID    int64
}

type CreateResponseInput struct {
	Actor     Actor
	VacancyID int64
	Message   *string
}

type WithdrawResponseInput struct {
	Actor      Actor
	ResponseID int64
}

type ListResponsesInput struct {
	Actor     Actor
	VacancyID int64
	PageSize  int32
	PageToken *string
}

type ListMyResponsesInput struct {
	Actor     Actor
	PageSize  int32
	PageToken *string
}

type AcceptResponseInput struct {
	Actor       Actor
	ResponseID  int64
	AdminReason *string
}

type ListMyAssignmentsInput struct {
	Actor     Actor
	PageSize  int32
	PageToken *string
}

type GetAssignmentForVacancyInput struct {
	Actor     Actor
	VacancyID int64
}

type CancelAssignmentInput struct {
	Actor        Actor
	AssignmentID int64
	Reason       *string
	AdminReason  *string
}

type CompleteAssignmentInput struct {
	Actor        Actor
	AssignmentID int64
	AdminReason  *string
}

type CreateReviewInput struct {
	Actor        Actor
	AssignmentID int64
	Rating       int32
	Body         *string
}

type UpdateReviewInput struct {
	Actor    Actor
	ReviewID int64
	Rating   *int32
	Body     *string
}

type CreateReviewReplyInput struct {
	Actor        Actor
	ReviewID     int64
	TargetUserID int64
	Body         string
}

type UpdateReviewReplyInput struct {
	Actor   Actor
	ReplyID int64
	Body    string
}

type ListReviewsInput struct {
	Actor        Actor
	TargetUserID *int64
	PageSize     int32
	PageToken    *string
}

type ListReviewRepliesInput struct {
	Actor     Actor
	ReviewID  int64
	PageSize  int32
	PageToken *string
}

type GetUserRatingInput struct {
	Actor  Actor
	UserID int64
}
