package dto

type EmailVerificationRequested struct {
	UserID int64
	Email  string
	Code   string
}

type ResponseCreated struct {
	ResponseID   int64
	VacancyID    int64
	OwnerUserID  int64
	OwnerEmail   string
	WorkerUserID int64
	VacancyTitle string
	ShiftDate    string
	PayAmount    int64
}
