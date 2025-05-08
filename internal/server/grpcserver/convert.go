package grpcserver

import (
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	schedulev1 "schedule/pkg/grpc"
)

func newDomainScheduleWithDuration(req *schedulev1.CreateScheduleRequest) *entity.ScheduleWithDuration {
	return &entity.ScheduleWithDuration{
		UserId:   value.UserId(req.GetUserId()),
		Name:     value.ScheduleName(req.GetName()),
		Duration: value.ScheduleDuration(req.GetDuration()),
		Period:   value.SchedulePeriod(req.GetPeriod()),
	}
}

func newGRPCCreateScheduleReply(scheduleId value.ScheduleId) *schedulev1.CreateScheduleReply {
	return &schedulev1.CreateScheduleReply{
		Id: int32(scheduleId),
	}
}

func newGRPCGetScheduleReply(timetable *entity.ScheduleTimetable) *schedulev1.GetScheduleReply {
	grpcTimetable := make([]int64, len(timetable.Timetable))
	for i, t := range timetable.Timetable {
		grpcTimetable[i] = t.Unix()
	}

	grpcResp := &schedulev1.GetScheduleReply{
		Name:      timetable.Name.String(),
		Period:    int64(timetable.Period),
		Timetable: grpcTimetable,
	}
	if !timetable.EndAt.IsNil() {
		grpcResp.EndAt = timetable.EndAt.Unix()
	}

	return grpcResp
}

func newGRPCGetSchedulesReply(ids []value.ScheduleId) *schedulev1.GetSchedulesReply {
	grpcIds := make([]int32, len(ids))
	for i, id := range ids {
		grpcIds[i] = int32(id)
	}

	return &schedulev1.GetSchedulesReply{
		ScheduleIds: grpcIds,
	}
}

func newGRPCGetNextTakingsReply(schedules []entity.ScheduleNextTaking) *schedulev1.GetNextTakingsReply {
	grpcRespItems := make([]*schedulev1.GetNextTakingsReplyItem, len(schedules))

	for i, item := range schedules {
		grpcRespItems[i] = &schedulev1.GetNextTakingsReplyItem{
			Id:         int32(item.Id),
			Name:       item.Name.String(),
			Period:     int64(item.Period),
			NextTaking: item.NextTaking.Unix(),
		}
		if !item.EndAt.IsNil() {
			grpcRespItems[i].EndAt = item.EndAt.Unix()
		}
	}
	return &schedulev1.GetNextTakingsReply{
		Items: grpcRespItems,
	}
}
