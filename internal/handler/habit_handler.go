package handler

import (
	"context"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/yourorg/habit-tracker/gen/habit/v1"
	"github.com/yourorg/habit-tracker/gen/habit/v1/habitv1connect"
	"github.com/yourorg/habit-tracker/internal/domain"
	"github.com/yourorg/habit-tracker/internal/middleware"
	"github.com/yourorg/habit-tracker/internal/service"
)

// HabitHandler implements habitv1connect.HabitServiceHandler.
type HabitHandler struct {
	svc *service.HabitService
}

var _ habitv1connect.HabitServiceHandler = (*HabitHandler)(nil)

func NewHabitHandler(svc *service.HabitService) *HabitHandler {
	return &HabitHandler{svc: svc}
}

func (h *HabitHandler) CreateHabit(
	ctx context.Context,
	req *connect.Request[pb.CreateHabitRequest],
) (*connect.Response[pb.CreateHabitResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	habit := &domain.Habit{
		Title:        req.Msg.Title,
		Color:        req.Msg.Color,
		TargetPerDay: int(req.Msg.TargetPerDay),
		ScheduleType: protoScheduleToDomain(req.Msg.ScheduleType),
		Weekdays:     protoWeekdaysToDomain(req.Msg.Weekdays),
	}
	created, err := h.svc.CreateHabit(ctx, userID, habit)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.CreateHabitResponse{Habit: habitToProto(created)}), nil
}

func (h *HabitHandler) UpdateHabit(
	ctx context.Context,
	req *connect.Request[pb.UpdateHabitRequest],
) (*connect.Response[pb.UpdateHabitResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	habit := &domain.Habit{
		ID:           req.Msg.Id,
		Title:        req.Msg.Title,
		Color:        req.Msg.Color,
		TargetPerDay: int(req.Msg.TargetPerDay),
		ScheduleType: protoScheduleToDomain(req.Msg.ScheduleType),
		Weekdays:     protoWeekdaysToDomain(req.Msg.Weekdays),
	}
	updated, err := h.svc.UpdateHabit(ctx, userID, habit)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.UpdateHabitResponse{Habit: habitToProto(updated)}), nil
}

func (h *HabitHandler) DeleteHabit(
	ctx context.Context,
	req *connect.Request[pb.DeleteHabitRequest],
) (*connect.Response[pb.DeleteHabitResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	if err := h.svc.DeleteHabit(ctx, userID, req.Msg.Id); err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.DeleteHabitResponse{}), nil
}

func (h *HabitHandler) GetHabit(
	ctx context.Context,
	req *connect.Request[pb.GetHabitRequest],
) (*connect.Response[pb.GetHabitResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	habit, err := h.svc.GetHabit(ctx, userID, req.Msg.Id)
	if err != nil {
		return nil, toConnectError(err)
	}
	return connect.NewResponse(&pb.GetHabitResponse{Habit: habitToProto(habit)}), nil
}

func (h *HabitHandler) ListHabits(
	ctx context.Context,
	req *connect.Request[pb.ListHabitsRequest],
) (*connect.Response[pb.ListHabitsResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	page := int(req.Msg.Page)
	if page < 1 {
		page = 1
	}
	size := int(req.Msg.PageSize)
	if size < 1 {
		size = 50
	}
	habits, total, err := h.svc.ListHabits(ctx, userID, page, size)
	if err != nil {
		return nil, toConnectError(err)
	}
	pbHabits := make([]*pb.Habit, len(habits))
	for i, h := range habits {
		pbHabits[i] = habitToProto(h)
	}
	return connect.NewResponse(&pb.ListHabitsResponse{
		Habits: pbHabits,
		Total:  int32(total),
	}), nil
}

func (h *HabitHandler) GetTodaysHabits(
	ctx context.Context,
	req *connect.Request[pb.GetTodaysHabitsRequest],
) (*connect.Response[pb.GetTodaysHabitsResponse], error) {
	userID, ok := middleware.UserIDFromContext(ctx)
	if !ok {
		return nil, errUnauth
	}
	views, err := h.svc.GetTodaysHabits(ctx, userID)
	if err != nil {
		return nil, toConnectError(err)
	}
	pbHabits := make([]*pb.TodayHabit, len(views))
	for i, v := range views {
		pbHabits[i] = &pb.TodayHabit{
			Habit:     habitToProto(v.Habit),
			Progress:  int32(v.Progress),
			Target:    int32(v.Target),
			Completed: v.Completed,
		}
	}
	return connect.NewResponse(&pb.GetTodaysHabitsResponse{Habits: pbHabits}), nil
}

// ── Mapping helpers ──────────────────────────────────────────────────────────

var errUnauth = connect.NewError(connect.CodeUnauthenticated, domain.ErrUnauthorized)

func habitToProto(h *domain.Habit) *pb.Habit {
	return &pb.Habit{
		Id:           h.ID,
		UserId:       h.UserID,
		Title:        h.Title,
		Color:        h.Color,
		TargetPerDay: int32(h.TargetPerDay),
		ScheduleType: domainScheduleToProto(h.ScheduleType),
		Weekdays:     domainWeekdaysToProto(h.Weekdays),
		CreatedAt:    timestamppb.New(h.CreatedAt),
		UpdatedAt:    timestamppb.New(h.UpdatedAt),
	}
}

func domainScheduleToProto(s domain.ScheduleType) pb.ScheduleType {
	switch s {
	case domain.ScheduleDaily:
		return pb.ScheduleType_SCHEDULE_TYPE_DAILY
	case domain.ScheduleEveryOtherDay:
		return pb.ScheduleType_SCHEDULE_TYPE_EVERY_OTHER_DAY
	case domain.ScheduleWeekdays:
		return pb.ScheduleType_SCHEDULE_TYPE_WEEKDAYS
	default:
		return pb.ScheduleType_SCHEDULE_TYPE_UNSPECIFIED
	}
}

func protoScheduleToDomain(s pb.ScheduleType) domain.ScheduleType {
	switch s {
	case pb.ScheduleType_SCHEDULE_TYPE_EVERY_OTHER_DAY:
		return domain.ScheduleEveryOtherDay
	case pb.ScheduleType_SCHEDULE_TYPE_WEEKDAYS:
		return domain.ScheduleWeekdays
	default:
		return domain.ScheduleDaily
	}
}

func domainWeekdaysToProto(wds []domain.Weekday) []pb.Weekday {
	out := make([]pb.Weekday, len(wds))
	for i, w := range wds {
		// Proto enum offset: 0=UNSPECIFIED, 1=Sunday…7=Saturday
		out[i] = pb.Weekday(int32(w) + 1)
	}
	return out
}

func protoWeekdaysToDomain(wds []pb.Weekday) []domain.Weekday {
	out := make([]domain.Weekday, 0, len(wds))
	for _, w := range wds {
		v := int(w) - 1
		if v >= 0 && v <= 6 {
			out = append(out, domain.Weekday(v))
		}
	}
	return out
}
