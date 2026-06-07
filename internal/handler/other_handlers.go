package handler

import (
	"context"

	"connectrpc.com/connect"

	pb "github.com/yourorg/habit-tracker/gen/habit/v1"
	"github.com/yourorg/habit-tracker/gen/habit/v1/habitv1connect"
	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/middleware"
	"github.com/yourorg/habit-tracker/internal/service"
)

// ─── Progress Handler ────────────────────────────────────────────────────────

type ProgressHandler struct {
	svc *service.ProgressService
}

var _ habitv1connect.ProgressServiceHandler = (*ProgressHandler)(nil)

func NewProgressHandler(svc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{svc: svc}
}

func (h *ProgressHandler) IncrementProgress(
	ctx context.Context,
	req *connect.Request[pb.IncrementProgressRequest],
) (*connect.Response[pb.ProgressResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	res, err := h.svc.IncrementProgress(ctx, userID, req.Msg.HabitId, req.Msg.Date)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(progressToProto(res)), nil
}

func (h *ProgressHandler) DecrementProgress(
	ctx context.Context,
	req *connect.Request[pb.DecrementProgressRequest],
) (*connect.Response[pb.ProgressResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	res, err := h.svc.DecrementProgress(ctx, userID, req.Msg.HabitId, req.Msg.Date)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(progressToProto(res)), nil
}

func (h *ProgressHandler) ResetProgress(
	ctx context.Context,
	req *connect.Request[pb.ResetProgressRequest],
) (*connect.Response[pb.ProgressResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	res, err := h.svc.ResetProgress(ctx, userID, req.Msg.HabitId, req.Msg.Date)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(progressToProto(res)), nil
}

func (h *ProgressHandler) GetProgress(
	ctx context.Context,
	req *connect.Request[pb.GetProgressRequest],
) (*connect.Response[pb.ProgressResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	res, err := h.svc.GetProgress(ctx, userID, req.Msg.HabitId, req.Msg.Date)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(progressToProto(res)), nil
}

func progressToProto(r *service.ProgressResult) *pb.ProgressResponse {
	return &pb.ProgressResponse{
		HabitId:   r.HabitID,
		Date:      r.Date,
		Progress:  int32(r.Progress),
		Target:    int32(r.Target),
		Completed: r.Completed,
	}
}

// ─── Settings Handler ────────────────────────────────────────────────────────

type SettingsHandler struct {
	svc *service.SettingsService
}

var _ habitv1connect.SettingsServiceHandler = (*SettingsHandler)(nil)

func NewSettingsHandler(svc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{svc: svc}
}

func (h *SettingsHandler) GetSettings(
	ctx context.Context,
	_ *connect.Request[pb.GetSettingsRequest],
) (*connect.Response[pb.GetSettingsResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	s, err := h.svc.GetSettings(ctx, userID)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.GetSettingsResponse{
		Settings: &pb.Settings{
			Language:             s.Language,
			NotificationsEnabled: s.NotificationsEnabled,
			ShowDates:            s.ShowDates,
		},
	}), nil
}

func (h *SettingsHandler) UpdateSettings(
	ctx context.Context,
	req *connect.Request[pb.UpdateSettingsRequest],
) (*connect.Response[pb.UpdateSettingsResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	in := &domain.Settings{
		Language:             req.Msg.Language,
		NotificationsEnabled: req.Msg.NotificationsEnabled,
		ShowDates:            req.Msg.ShowDates,
	}
	s, err := h.svc.UpdateSettings(ctx, userID, in)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.UpdateSettingsResponse{
		Settings: &pb.Settings{
			Language:             s.Language,
			NotificationsEnabled: s.NotificationsEnabled,
			ShowDates:            s.ShowDates,
		},
	}), nil
}

// ─── Stats Handler ───────────────────────────────────────────────────────────

type StatsHandler struct {
	svc *service.StatsService
}

var _ habitv1connect.StatsServiceHandler = (*StatsHandler)(nil)

func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

func (h *StatsHandler) GetStats(
	ctx context.Context,
	req *connect.Request[pb.GetStatsRequest],
) (*connect.Response[pb.GetStatsResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	stats, err := h.svc.GetStats(ctx, userID, req.Msg.HabitId)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.GetStatsResponse{
		CurrentStreak:      int32(stats.CurrentStreak),
		BestStreak:         int32(stats.BestStreak),
		TrackedHabitsCount: int32(stats.TrackedHabitsCount),
		CompletionRate_30D: stats.CompletionRate30d,
	}), nil
}
