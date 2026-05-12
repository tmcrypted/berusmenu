package store

import (
	"time"

	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
)

type Vacancy struct {
	ID                 int64
	OwnerID            int64
	PvzID              int64
	Status             commonv1.VacancyStatus
	Title              *string
	Description        *string
	ShiftStartAt       time.Time
	ShiftEndAt         time.Time
	PvzNameSnapshot    string
	PvzAddressSnapshot string
	PvzCitySnapshot    string
	PayAmount          int64
	SlotsTotal         int32
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Response struct {
	ID        int64
	VacancyID int64
	WorkerID  int64
	Status    commonv1.ResponseStatus
	Message   *string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Assignment struct {
	ID                   int64
	VacancyID            int64
	ResponseID           int64
	OwnerID              int64
	WorkerID             int64
	Status               commonv1.AssignmentStatus
	AcceptedAt           time.Time
	PvzNameSnapshot      string
	PvzAddressSnapshot   string
	PvzCitySnapshot      string
	ShiftStartAtSnapshot time.Time
	PayAmountSnapshot    int64
	ShiftEndAtSnapshot   time.Time
	OwnerNameSnapshot    string
	WorkerNameSnapshot   string
	CompletedAt          *time.Time
	CancelledAt          *time.Time
	CancelledByUserID    *int64
	CancelReason         *string
	UpdatedAt            time.Time
	CreatedAt            time.Time
}

type Review struct {
	ID           int64
	AssignmentID int64
	AuthorUserID int64
	TargetUserID int64
	TargetRole   commonv1.ReviewTargetRole
	Rating       int32
	Body         *string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type ReviewReply struct {
	ID           int64
	ReviewID     int64
	AuthorUserID int64
	TargetUserID int64
	Body         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type UserRating struct {
	UserID       int64
	AvgRating    float64
	ReviewsCount int64
	UpdatedAt    time.Time
}

type OutboxEvent struct {
	ID        int64
	EventID   string
	Topic     string
	EventType string
	EventKey  string
	Payload   []byte
	Attempts  int
}

type OutboxEventInput struct {
	EventID   string
	Topic     string
	EventType string
	EventKey  string
	Payload   []byte
}

type AdminEventInput struct {
	ActorUserID   int64
	TargetType    string
	TargetID      int64
	Action        string
	PreviousState *string
	NewState      *string
	Reason        string
	Comment       *string
}

type AssignmentSnapshotInput struct {
	ResponseID         int64
	Vacancy            Vacancy
	WorkerID           int64
	PvzNameSnapshot    string
	PvzAddressSnapshot string
	PvzCitySnapshot    string
	OwnerNameSnapshot  string
	WorkerNameSnapshot string
}
