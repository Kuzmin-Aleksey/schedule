package grpcServer

import (
	"schedule/internal/entity"
	schedulev1 "schedule/internal/server/grpcServer/gen"
	"schedule/internal/value"
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

func newGRPCGetTimetableReply(timetable *entity.ScheduleTimetable) *schedulev1.GetTimetableReply {
	grpcTimetable := make([]int64, len(timetable.Timetable))
	for i, t := range timetable.Timetable {
		grpcTimetable[i] = t.Unix()
	}

	grpcResp := &schedulev1.GetTimetableReply{
		Name:      timetable.Name.String(),
		Period:    int64(timetable.Period),
		Timetable: grpcTimetable,
	}
	if !timetable.EndAt.IsNil() {
		grpcResp.EndAt = timetable.EndAt.Unix()
	}

	return grpcResp
}

func newGRPCGetByUserReply(ids []value.ScheduleId) *schedulev1.GetByUserReply {
	grpcIds := make([]int32, len(ids))
	for i, id := range ids {
		grpcIds[i] = int32(id)
	}

	return &schedulev1.GetByUserReply{
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
