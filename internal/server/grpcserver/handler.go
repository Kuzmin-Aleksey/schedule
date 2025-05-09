package grpcserver

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
	"schedule/internal/domain/value"
	"schedule/internal/server"
	"schedule/pkg/contextx"
	schedulev1 "schedule/pkg/grpc"
)

type scheduleAPI struct {
	schedulev1.ScheduleServer
	schedule server.ScheduleUsecase
}

func Register(server *grpc.Server, schedule server.ScheduleUsecase) {
	schedulev1.RegisterScheduleServer(server, &scheduleAPI{
		schedule: schedule,
	})
}

func (s *scheduleAPI) CreateSchedule(ctx context.Context, req *schedulev1.CreateScheduleRequest) (*schedulev1.CreateScheduleReply, error) {
	l := contextx.GetLoggerOrDefault(ctx)

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
		l.LogAttrs(ctx, slog.LevelError, "handling request error", slog.String("err", err.Error()))
		return nil, status.Error(getCodeFromError(err), "create schedule error")
	}

	return newGRPCCreateScheduleReply(resp), nil
}

func (s *scheduleAPI) GetSchedule(ctx context.Context, req *schedulev1.GetScheduleRequest) (*schedulev1.GetScheduleReply, error) {
	l := contextx.GetLoggerOrDefault(ctx)

	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.GetScheduleId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "schedule id is required")
	}

	resp, err := s.schedule.GetTimetable(ctx, value.UserId(req.GetUserId()), value.ScheduleId(req.GetScheduleId()))
	if err != nil {
		l.LogAttrs(ctx, slog.LevelError, "handling request error", slog.String("err", err.Error()))
		return nil, status.Error(getCodeFromError(err), "get schedule error")
	}

	return newGRPCGetScheduleReply(resp), nil
}

func (s *scheduleAPI) GetSchedules(ctx context.Context, req *schedulev1.GetSchedulesRequest) (*schedulev1.GetSchedulesReply, error) {
	l := contextx.GetLoggerOrDefault(ctx)

	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	ids, err := s.schedule.GetByUser(ctx, value.UserId(req.GetUserId()))
	if err != nil {
		l.LogAttrs(ctx, slog.LevelError, "handling request error", slog.String("err", err.Error()))
		return nil, status.Error(getCodeFromError(err), "get schedules error")
	}

	return newGRPCGetSchedulesReply(ids), nil
}

func (s *scheduleAPI) GetNextTakings(ctx context.Context, req *schedulev1.GetNextTakingsRequest) (*schedulev1.GetNextTakingsReply, error) {
	l := contextx.GetLoggerOrDefault(ctx)

	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	nextTakings, err := s.schedule.GetNextTakings(ctx, value.UserId(req.GetUserId()))
	if err != nil {
		l.ErrorContext(ctx, "handling request error", "err", err)
		return nil, status.Error(getCodeFromError(err), "get next takings error")
	}

	return newGRPCGetNextTakingsReply(nextTakings), nil
}
