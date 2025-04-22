package grpcHandler

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"schedule/gen-proto"
	"schedule/internal/usecase/schedule"
	"schedule/pkg/logger"
	"time"
)

type scheduleAPI struct {
	schedulev1.UnimplementedScheduleServer
	schedule *schedule.Usecase
	l        *logger.Logger
}

func Register(server *grpc.Server, schedule *schedule.Usecase, l *logger.Logger) {
	schedulev1.RegisterScheduleServer(server, &scheduleAPI{
		schedule: schedule,
		l:        l,
	})
}

const errInternal = "internal error"

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

	resp, err := s.schedule.Create(ctx, &schedule.CreateScheduleDTO{
		UserId:   req.GetUserId(),
		Name:     req.GetName(),
		Duration: uint(req.GetDuration()),
		Period:   time.Duration(req.GetPeriod()),
	})
	if err != nil {
		s.l.Error(err)
		return nil, status.Error(codes.Internal, errInternal)
	}

	return &schedulev1.CreateScheduleReply{
		Id: int32(resp.Id),
	}, nil
}

func (s *scheduleAPI) GetTimetable(ctx context.Context, req *schedulev1.GetTimetableRequest) (*schedulev1.GetTimetableReply, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}
	if req.GetScheduleId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "schedule id is required")
	}

	resp, err := s.schedule.GetTimetable(ctx, req.GetUserId(), int(req.GetScheduleId()))
	if err != nil {
		s.l.Error(err)
		return nil, status.Error(codes.Internal, errInternal)
	}

	grpcTimetable := make([]int64, len(resp.Timetable))
	for i, t := range resp.Timetable {
		grpcTimetable[i] = t.Unix()
	}

	grpcResp := &schedulev1.GetTimetableReply{
		Name:      resp.Name,
		Period:    int64(resp.Period),
		Timetable: grpcTimetable,
	}
	if resp.EndAt != nil {
		grpcResp.EndAt = resp.EndAt.Unix()
	}

	return grpcResp, nil
}

func (s *scheduleAPI) GetByUser(ctx context.Context, req *schedulev1.GetByUserRequest) (*schedulev1.GetByUserReply, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	ids, err := s.schedule.GetByUser(ctx, req.GetUserId())
	if err != nil {
		s.l.Error(err)
		return nil, status.Error(codes.Internal, errInternal)
	}

	grpcIds := make([]int32, len(ids))
	for i, id := range ids {
		grpcIds[i] = int32(id)
	}

	return &schedulev1.GetByUserReply{
		ScheduleIds: grpcIds,
	}, nil
}

func (s *scheduleAPI) GetNextTakings(ctx context.Context, req *schedulev1.GetNextTakingsRequest) (*schedulev1.GetNextTakingsReply, error) {
	if req.GetUserId() == 0 {
		return nil, status.Error(codes.InvalidArgument, "user id is required")
	}

	resp, err := s.schedule.GetNextTakings(ctx, req.GetUserId())
	if err != nil {
		s.l.Error(err)
		return nil, status.Error(codes.Internal, errInternal)
	}

	grpcRespItems := make([]*schedulev1.GetNextTakingsReplyItem, len(resp))
	for i, item := range resp {
		grpcRespItems[i] = &schedulev1.GetNextTakingsReplyItem{
			Id:         int32(item.Id),
			Name:       item.Name,
			Period:     int64(item.Period),
			NextTaking: item.NextTaking.Unix(),
		}
		if item.EndAt != nil {
			grpcRespItems[i].EndAt = item.EndAt.Unix()
		}
	}

	return &schedulev1.GetNextTakingsReply{
		Items: grpcRespItems,
	}, nil
}
