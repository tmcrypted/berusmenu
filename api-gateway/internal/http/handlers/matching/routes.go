package matching

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/dto"
	"github.com/tmcrypted/berusmenu/api-gateway/internal/http/response"
	commonv1 "github.com/tmcrypted/berusmenu/proto/gen/go/common/v1"
	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Handler struct {
	matching matchingv1.MatchingServiceClient
}

func RegisterRoutes(
	group *gin.RouterGroup,
	client matchingv1.MatchingServiceClient,
	authMiddleware gin.HandlerFunc,
	csrfMiddleware gin.HandlerFunc,
	verifiedMiddleware gin.HandlerFunc,
	moderatorMiddleware gin.HandlerFunc,
) {
	handler := Handler{matching: client}

	authorized := group.Group("", authMiddleware, verifiedMiddleware)
	authorized.POST("/vacancies", csrfMiddleware, handler.CreateVacancy)
	authorized.GET("/vacancies", handler.ListVacancies)
	authorized.GET("/vacancies/:id", handler.GetVacancy)
	authorized.POST("/vacancies/:id/responses", csrfMiddleware, handler.CreateResponse)
	authorized.GET("/vacancies/:id/responses", handler.ListResponsesForVacancy)
	authorized.POST("/responses/:id/withdraw", csrfMiddleware, handler.WithdrawResponse)
	authorized.POST("/responses/:id/accept", csrfMiddleware, handler.AcceptResponse)
	authorized.POST("/assignments/:id/cancel", csrfMiddleware, handler.CancelAssignment)
	authorized.POST("/assignments/:id/complete", csrfMiddleware, handler.CompleteAssignment)
	authorized.POST("/assignments/:id/reviews", csrfMiddleware, handler.CreateReview)
	authorized.PATCH("/reviews/:id", csrfMiddleware, handler.UpdateReview)
	authorized.GET("/reviews", handler.ListReviews)
	authorized.POST("/reviews/:id/replies", csrfMiddleware, handler.CreateReviewReply)
	authorized.GET("/reviews/:id/replies", handler.ListReviewReplies)
	authorized.PATCH("/review-replies/:id", csrfMiddleware, handler.UpdateReviewReply)
	authorized.GET("/users/:id/rating", handler.GetUserRating)

	moderation := authorized.Group("/moderation/matching", moderatorMiddleware)
	moderation.GET("/vacancies", handler.ListVacancies)
	moderation.POST("/responses/:id/accept", csrfMiddleware, handler.ModerationAcceptResponse)
	moderation.POST("/assignments/:id/cancel", csrfMiddleware, handler.ModerationCancelAssignment)
	moderation.POST("/assignments/:id/complete", csrfMiddleware, handler.ModerationCompleteAssignment)
}

func (h Handler) CreateVacancy(c *gin.Context) {
	var req dto.CreateVacancyRequest
	if !bindJSON(c, &req) {
		return
	}
	start, ok := parseTime(c, req.ShiftStartAt, "shift_start_at")
	if !ok {
		return
	}
	end, ok := parseTime(c, req.ShiftEndAt, "shift_end_at")
	if !ok {
		return
	}
	res, err := h.matching.CreateVacancy(c.Request.Context(), &matchingv1.CreateVacancyRequest{
		PvzId:        req.PvzID,
		Title:        req.Title,
		Description:  req.Description,
		ShiftStartAt: timestamppb.New(start),
		ShiftEndAt:   timestamppb.New(end),
		PayAmount:    req.PayAmount,
		SlotsTotal:   req.SlotsTotal,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusCreated, res.GetVacancy())
}

func (h Handler) ListVacancies(c *gin.Context) {
	pageSize, ok := pageSize(c)
	if !ok {
		return
	}
	req := &matchingv1.ListVacanciesRequest{PageSize: pageSize}
	if token := c.Query("page_token"); token != "" {
		req.PageToken = &token
	}
	if raw := c.Query("pvz_id"); raw != "" {
		id, ok := parsePositiveInt(c, raw, "pvz_id")
		if !ok {
			return
		}
		req.PvzId = &id
	}
	if raw := c.Query("status"); raw != "" {
		status, ok := vacancyStatus(c, raw)
		if !ok {
			return
		}
		req.Status = &status
	}
	if raw := c.Query("shift_start_from"); raw != "" {
		value, ok := parseTime(c, raw, "shift_start_from")
		if !ok {
			return
		}
		req.ShiftStartFrom = timestamppb.New(value)
	}
	if raw := c.Query("shift_start_to"); raw != "" {
		value, ok := parseTime(c, raw, "shift_start_to")
		if !ok {
			return
		}
		req.ShiftStartTo = timestamppb.New(value)
	}
	res, err := h.matching.ListVacancies(c.Request.Context(), req)
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res)
}

func (h Handler) GetVacancy(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	res, err := h.matching.GetVacancy(c.Request.Context(), &matchingv1.GetVacancyRequest{Id: id})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetVacancy())
}

func (h Handler) CreateResponse(c *gin.Context) {
	vacancyID, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.CreateResponseRequest
	if !bindOptionalJSON(c, &req) {
		return
	}
	res, err := h.matching.CreateResponse(c.Request.Context(), &matchingv1.CreateResponseRequest{VacancyId: vacancyID, Message: req.Message})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusCreated, res.GetResponse())
}

func (h Handler) ListResponsesForVacancy(c *gin.Context) {
	vacancyID, ok := pathID(c, "id")
	if !ok {
		return
	}
	pageSize, ok := pageSize(c)
	if !ok {
		return
	}
	req := &matchingv1.ListResponsesForVacancyRequest{VacancyId: vacancyID, PageSize: pageSize}
	if token := c.Query("page_token"); token != "" {
		req.PageToken = &token
	}
	res, err := h.matching.ListResponsesForVacancy(c.Request.Context(), req)
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res)
}

func (h Handler) WithdrawResponse(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	res, err := h.matching.WithdrawResponse(c.Request.Context(), &matchingv1.WithdrawResponseRequest{ResponseId: id})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetResponse())
}

func (h Handler) AcceptResponse(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	res, err := h.matching.AcceptResponse(c.Request.Context(), &matchingv1.AcceptResponseRequest{ResponseId: id})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetAssignment())
}

func (h Handler) ModerationAcceptResponse(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.ModeratorActionRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.matching.AcceptResponse(c.Request.Context(), &matchingv1.AcceptResponseRequest{ResponseId: id, AdminReason: &req.Reason})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetAssignment())
}

func (h Handler) CancelAssignment(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.CancelAssignmentRequest
	if !bindOptionalJSON(c, &req) {
		return
	}
	res, err := h.matching.CancelAssignment(c.Request.Context(), &matchingv1.CancelAssignmentRequest{AssignmentId: id, CancelReason: req.CancelReason})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetAssignment())
}

func (h Handler) ModerationCancelAssignment(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.ModeratorCancelAssignmentRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.matching.CancelAssignment(c.Request.Context(), &matchingv1.CancelAssignmentRequest{
		AssignmentId: id,
		CancelReason: req.CancelReason,
		AdminReason:  &req.Reason,
	})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetAssignment())
}

func (h Handler) CompleteAssignment(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	res, err := h.matching.CompleteAssignment(c.Request.Context(), &matchingv1.CompleteAssignmentRequest{AssignmentId: id})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetAssignment())
}

func (h Handler) ModerationCompleteAssignment(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.ModeratorActionRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.matching.CompleteAssignment(c.Request.Context(), &matchingv1.CompleteAssignmentRequest{AssignmentId: id, AdminReason: &req.Reason})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetAssignment())
}

func (h Handler) CreateReview(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.CreateReviewRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.matching.CreateReview(c.Request.Context(), &matchingv1.CreateReviewRequest{AssignmentId: id, Rating: req.Rating, Body: req.Body})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusCreated, res.GetReview())
}

func (h Handler) UpdateReview(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateReviewRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.matching.UpdateReview(c.Request.Context(), &matchingv1.UpdateReviewRequest{ReviewId: id, Rating: req.Rating, Body: req.Body})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetReview())
}

func (h Handler) ListReviews(c *gin.Context) {
	pageSize, ok := pageSize(c)
	if !ok {
		return
	}
	req := &matchingv1.ListReviewsRequest{PageSize: pageSize}
	if token := c.Query("page_token"); token != "" {
		req.PageToken = &token
	}
	if raw := c.Query("target_user_id"); raw != "" {
		id, ok := parsePositiveInt(c, raw, "target_user_id")
		if !ok {
			return
		}
		req.TargetUserId = &id
	}
	res, err := h.matching.ListReviews(c.Request.Context(), req)
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res)
}

func (h Handler) ListReviewReplies(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	pageSize, ok := pageSize(c)
	if !ok {
		return
	}
	req := &matchingv1.ListReviewRepliesRequest{ReviewId: id, PageSize: pageSize}
	if token := c.Query("page_token"); token != "" {
		req.PageToken = &token
	}
	res, err := h.matching.ListReviewReplies(c.Request.Context(), req)
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res)
}

func (h Handler) GetUserRating(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	res, err := h.matching.GetUserRating(c.Request.Context(), &matchingv1.GetUserRatingRequest{UserId: id})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetRating())
}

func (h Handler) CreateReviewReply(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.CreateReviewReplyRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.matching.CreateReviewReply(c.Request.Context(), &matchingv1.CreateReviewReplyRequest{ReviewId: id, TargetUserId: req.TargetUserID, Body: req.Body})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusCreated, res.GetReply())
}

func (h Handler) UpdateReviewReply(c *gin.Context) {
	id, ok := pathID(c, "id")
	if !ok {
		return
	}
	var req dto.UpdateReviewReplyRequest
	if !bindJSON(c, &req) {
		return
	}
	res, err := h.matching.UpdateReviewReply(c.Request.Context(), &matchingv1.UpdateReviewReplyRequest{ReplyId: id, Body: req.Body})
	if err != nil {
		response.FromGRPCError(c, err)
		return
	}
	response.Proto(c, http.StatusOK, res.GetReply())
}

func bindJSON(c *gin.Context, dst any) bool {
	if err := c.ShouldBindJSON(dst); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_argument", err.Error())
		return false
	}
	return true
}

func bindOptionalJSON(c *gin.Context, dst any) bool {
	if c.Request.Body == nil || c.Request.ContentLength == 0 {
		return true
	}
	return bindJSON(c, dst)
}

func pathID(c *gin.Context, name string) (int64, bool) {
	return parsePositiveInt(c, c.Param(name), name)
}

func parsePositiveInt(c *gin.Context, raw string, name string) (int64, bool) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		response.Error(c, http.StatusBadRequest, "invalid_argument", "invalid "+name)
		return 0, false
	}
	return id, true
}

func pageSize(c *gin.Context) (int32, bool) {
	raw := c.Query("page_size")
	if raw == "" {
		return 0, true
	}
	value, err := strconv.ParseInt(raw, 10, 32)
	if err != nil || value < 0 {
		response.Error(c, http.StatusBadRequest, "invalid_argument", "invalid page_size")
		return 0, false
	}
	return int32(value), true
}

func parseTime(c *gin.Context, raw string, field string) (time.Time, bool) {
	value, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		response.Error(c, http.StatusBadRequest, "invalid_argument", "invalid "+field)
		return time.Time{}, false
	}
	return value, true
}

func vacancyStatus(c *gin.Context, value string) (commonv1.VacancyStatus, bool) {
	switch value {
	case "assigned":
		return commonv1.VacancyStatus_VACANCY_STATUS_ASSIGNED, true
	case "completed":
		return commonv1.VacancyStatus_VACANCY_STATUS_COMPLETED, true
	case "cancelled":
		return commonv1.VacancyStatus_VACANCY_STATUS_CANCELLED, true
	case "published":
		return commonv1.VacancyStatus_VACANCY_STATUS_PUBLISHED, true
	default:
		response.Error(c, http.StatusBadRequest, "invalid_argument", "invalid status")
		return commonv1.VacancyStatus_VACANCY_STATUS_UNSPECIFIED, false
	}
}
