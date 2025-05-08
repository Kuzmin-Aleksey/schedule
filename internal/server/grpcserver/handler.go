package grpcserver

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"schedule/internal/domain/usecase/schedule"
	"schedule/internal/domain/value"
	schedulev1 "schedule/pkg/grpc"
)

type scheduleAPI struct {
	schedulev1.ScheduleServer
	schedule *schedule.Usecase
	l        *slog.Logger
}

func Register(server *grpc.Server, schedule *schedule.Usecase, l *slog.Logger) {
	schedulev1.RegisterScheduleServer(server, &scheduleAPI{
		schedule: schedule,
		l:        l,
	})
}

var errInternal = status.Error(codes.Internal, "internal error")

func (s *scheduleAPI) CreateSchedule(ctx context.Context, req *schedulev1.CreateScheduleRequest) (*schedulev1.CreateScheduleReply, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.GetName() == "" {
		return nil, status.Error(codes.InvalidArgument, "name is required")
	}
	if req.GetPeriod() == 0 {
		return nil, status.Error(codes.InvalidArgument, "period is required")
	}

	resp, err := s.schedule.Create(ctx, newDomainScheduleWithDuration(req))
	if err != nil {
		s.l.LogAttrs(ctx, slog.LevelError, "handling request error", slog.String("err", err.Error()))
		return nil, errInternal
	}

	return newGRPCCreateScheduleReply(resp), nil
}

func (s *scheduleAPI) GetSchedule(ctx context.Context, req *schedulev1.GetScheduleRequest) (*schedulev1.GetScheduleReply, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.GetScheduleId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "schedule id is required")
	}

	resp, err := s.schedule.GetTimetable(ctx, value.UserId(req.GetUserId()), value.ScheduleId(req.GetScheduleId()))
	if err != nil {
		s.l.LogAttrs(ctx, slog.LevelError, "handling request error", slog.String("err", err.Error()))
		return nil, errInternal
	}

	return newGRPCGetScheduleReply(resp), nil
}

func (s *scheduleAPI) GetSchedules(ctx context.Context, req *schedulev1.GetSchedulesRequest) (*schedulev1.GetSchedulesReply, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	ids, err := s.schedule.GetByUser(ctx, value.UserId(req.GetUserId()))
	if err != nil {
		s.l.LogAttrs(ctx, slog.LevelError, "handling request error", slog.String("err", err.Error()))
		return nil, errInternal
	}

	return newGRPCGetSchedulesReply(ids), nil
}

func (s *scheduleAPI) GetNextTakings(ctx context.Context, req *schedulev1.GetNextTakingsRequest) (*schedulev1.GetNextTakingsReply, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	nextTakings, err := s.schedule.GetNextTakings(ctx, value.UserId(req.GetUserId()))
	if err != nil {
		s.l.ErrorContext(ctx, "GetNextTakings failed", "err", err)
		return nil, errInternal
	}

	return newGRPCGetNextTakingsReply(nextTakings), nil
}
