package dto

type CreatePvzRequest struct {
	City            string   `json:"city" binding:"required"`
	Address         string   `json:"address" binding:"required"`
	CompanyName     string   `json:"company_name" binding:"required"`
	ContactPhone    string   `json:"contact_phone" binding:"required"`
	ContactTelegram *string  `json:"contact_telegram"`
	Description     *string  `json:"description"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
}

type UpdatePvzRequest struct {
	City            *string  `json:"city"`
	Address         *string  `json:"address"`
	CompanyName     *string  `json:"company_name"`
	ContactPhone    *string  `json:"contact_phone"`
	ContactTelegram *string  `json:"contact_telegram"`
	Description     *string  `json:"description"`
	Lat             *float64 `json:"lat"`
	Lng             *float64 `json:"lng"`
}

type RejectPvzRequest struct {
	Reason  string  `json:"reason" binding:"required"`
	Comment *string `json:"comment"`
}

type ModerationCommentRequest struct {
	Comment *string `json:"comment"`
}

type BlockUserAndDeletePvzRequest struct {
	Reason               string  `json:"reason" binding:"required"`
	BlockDurationMinutes int64   `json:"block_duration_minutes" binding:"required,min=1"`
	Comment              *string `json:"comment"`
}
