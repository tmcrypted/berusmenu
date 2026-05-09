package dto

type CreateVacancyRequest struct {
	PvzID        int64   `json:"pvz_id" binding:"required,min=1"`
	Title        *string `json:"title"`
	Description  *string `json:"description"`
	ShiftStartAt string  `json:"shift_start_at" binding:"required"`
	ShiftEndAt   string  `json:"shift_end_at" binding:"required"`
	PayAmount    int64   `json:"pay_amount" binding:"required,min=1"`
	SlotsTotal   int32   `json:"slots_total" binding:"required,min=1"`
}

type CreateResponseRequest struct {
	Message *string `json:"message"`
}

type CancelAssignmentRequest struct {
	CancelReason *string `json:"cancel_reason"`
}

type ModeratorActionRequest struct {
	Reason  string  `json:"reason" binding:"required"`
	Comment *string `json:"comment"`
}

type ModeratorCancelAssignmentRequest struct {
	Reason       string  `json:"reason" binding:"required"`
	CancelReason *string `json:"cancel_reason"`
	Comment      *string `json:"comment"`
}

type CreateReviewRequest struct {
	Rating int32   `json:"rating" binding:"required,min=1,max=5"`
	Body   *string `json:"body"`
}

type UpdateReviewRequest struct {
	Rating *int32  `json:"rating"`
	Body   *string `json:"body"`
}

type CreateReviewReplyRequest struct {
	TargetUserID int64  `json:"target_user_id" binding:"required,min=1"`
	Body         string `json:"body" binding:"required"`
}

type UpdateReviewReplyRequest struct {
	Body string `json:"body" binding:"required"`
}
