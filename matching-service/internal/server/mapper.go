package server

import (
	"github.com/tmcrypted/berusmenu/matching-service/internal/store"
	matchingv1 "github.com/tmcrypted/berusmenu/proto/gen/go/matching/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func toProtoVacancy(item store.Vacancy) *matchingv1.Vacancy {
	return &matchingv1.Vacancy{
		Id:                 item.ID,
		OwnerId:            item.OwnerID,
		PvzId:              item.PvzID,
		PvzNameSnapshot:    item.PvzNameSnapshot,
		Status:             item.Status,
		Title:              item.Title,
		Description:        item.Description,
		ShiftStartAt:       timestamppb.New(item.ShiftStartAt),
		ShiftEndAt:         timestamppb.New(item.ShiftEndAt),
		PvzAddressSnapshot: item.PvzAddressSnapshot,
		PayAmount:          item.PayAmount,
		SlotsTotal:         item.SlotsTotal,
		CreatedAt:          timestamppb.New(item.CreatedAt),
		UpdatedAt:          timestamppb.New(item.UpdatedAt),
	}
}

func toProtoVacancies(items []store.Vacancy) []*matchingv1.Vacancy {
	out := make([]*matchingv1.Vacancy, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoVacancy(item))
	}
	return out
}

func toProtoResponse(item store.Response) *matchingv1.Response {
	return &matchingv1.Response{
		Id:        item.ID,
		VacancyId: item.VacancyID,
		WorkerId:  item.WorkerID,
		Status:    item.Status,
		Message:   item.Message,
		CreatedAt: timestamppb.New(item.CreatedAt),
		UpdatedAt: timestamppb.New(item.UpdatedAt),
	}
}

func toProtoResponses(items []store.Response) []*matchingv1.Response {
	out := make([]*matchingv1.Response, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoResponse(item))
	}
	return out
}

func toProtoAssignment(item store.Assignment) *matchingv1.Assignment {
	out := &matchingv1.Assignment{
		Id:                   item.ID,
		VacancyId:            item.VacancyID,
		ResponseId:           item.ResponseID,
		OwnerId:              item.OwnerID,
		WorkerId:             item.WorkerID,
		Status:               item.Status,
		AcceptedAt:           timestamppb.New(item.AcceptedAt),
		PvzNameSnapshot:      item.PvzNameSnapshot,
		PvzAddressSnapshot:   item.PvzAddressSnapshot,
		ShiftStartAtSnapshot: timestamppb.New(item.ShiftStartAtSnapshot),
		PayAmountSnapshot:    item.PayAmountSnapshot,
		ShiftEndAtSnapshot:   timestamppb.New(item.ShiftEndAtSnapshot),
		OwnerNameSnapshot:    item.OwnerNameSnapshot,
		WorkerNameSnapshot:   item.WorkerNameSnapshot,
		CancelledByUserId:    item.CancelledByUserID,
		CancelReason:         item.CancelReason,
		UpdatedAt:            timestamppb.New(item.UpdatedAt),
		CreatedAt:            timestamppb.New(item.CreatedAt),
	}
	if item.CompletedAt != nil {
		out.CompletedAt = timestamppb.New(*item.CompletedAt)
	}
	if item.CancelledAt != nil {
		out.CancelledAt = timestamppb.New(*item.CancelledAt)
	}
	return out
}

func toProtoAssignments(items []store.Assignment) []*matchingv1.Assignment {
	out := make([]*matchingv1.Assignment, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoAssignment(item))
	}
	return out
}

func toProtoReview(item store.Review) *matchingv1.Review {
	return &matchingv1.Review{
		Id:           item.ID,
		AssignmentId: item.AssignmentID,
		AuthorUserId: item.AuthorUserID,
		TargetUserId: item.TargetUserID,
		TargetRole:   item.TargetRole,
		Rating:       item.Rating,
		Body:         item.Body,
		CreatedAt:    timestamppb.New(item.CreatedAt),
		UpdatedAt:    timestamppb.New(item.UpdatedAt),
	}
}

func toProtoReviews(items []store.Review) []*matchingv1.Review {
	out := make([]*matchingv1.Review, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoReview(item))
	}
	return out
}

func toProtoReviewReply(item store.ReviewReply) *matchingv1.ReviewReply {
	return &matchingv1.ReviewReply{
		Id:           item.ID,
		ReviewId:     item.ReviewID,
		AuthorUserId: item.AuthorUserID,
		TargetUserId: item.TargetUserID,
		Body:         item.Body,
		CreatedAt:    timestamppb.New(item.CreatedAt),
		UpdatedAt:    timestamppb.New(item.UpdatedAt),
	}
}

func toProtoReviewReplies(items []store.ReviewReply) []*matchingv1.ReviewReply {
	out := make([]*matchingv1.ReviewReply, 0, len(items))
	for _, item := range items {
		out = append(out, toProtoReviewReply(item))
	}
	return out
}

func toProtoUserRating(item store.UserRating) *matchingv1.UserRating {
	return &matchingv1.UserRating{
		UserId:       item.UserID,
		AvgRating:    item.AvgRating,
		ReviewsCount: item.ReviewsCount,
		UpdatedAt:    timestamppb.New(item.UpdatedAt),
	}
}
