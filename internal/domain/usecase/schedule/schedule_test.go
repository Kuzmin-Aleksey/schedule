package schedule

import (
	"bou.ke/monkey"
	"context"
	"fmt"
	"github.com/stretchr/testify/require"
	"log/slog"
	"schedule/internal/app/logger"
	"schedule/internal/config"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/internal/util"
	"testing"
	"time"
)

var testConfig = config.ScheduleConfig{
	NextTakingPeriod: time.Hour,
	BeginDayHour:     8,
	EndDayHour:       22,
	TimeRound:        time.Minute * 15,
}

var logConfig = config.LogConfig{
	Level:  slog.LevelDebug.String(),
	Format: "json",
}

var l *slog.Logger

const testUser value.UserId = 1234567890123456

var testSchedules []entity.Schedule

func init() {
	now := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return now })

	time.Local = nil

	var err error
	l, err = logger.GetLogger(&logConfig)
	if err != nil {
		panic(err)
	}
	testSchedules = []entity.Schedule{
		{
			Id:     1,
			UserId: testUser,
			Name:   "Test Schedule 1",
			EndAt:  value.NewScheduleEndAt(util.Ptr(date().Add(day))),
			Period: value.SchedulePeriod(time.Hour),
		},
		{
			Id:     2,
			UserId: testUser,
			Name:   "Test Schedule 2",
			EndAt:  value.NewScheduleEndAt(util.Ptr(date())),
			Period: value.SchedulePeriod(time.Hour * 12),
		},
		{
			Id:     3,
			UserId: testUser,
			Name:   "Test Schedule 3",
			EndAt:  value.NewScheduleEndAt(util.Ptr(date().Add(-day))),
			Period: value.SchedulePeriod(time.Hour),
		},
		{
			Id:     4,
			UserId: testUser,
			Name:   "Test Schedule 4",
			EndAt:  value.NewScheduleEndAt(nil),
			Period: value.SchedulePeriod(time.Hour + time.Minute*2),
		},
		{
			Id:     5,
			UserId: testUser,
			Name:   "Test Schedule 5",
			EndAt:  value.NewScheduleEndAt(util.Ptr(date().Add(day))),
			Period: value.SchedulePeriod(time.Duration(testConfig.EndDayHour-testConfig.BeginDayHour) * time.Hour),
		},
	}
}

func TestGetActualSchedulesIds(t *testing.T) {
	uc := NewUsecase(nil, l, testConfig)

	testCases := []struct {
		Location *time.Location
		Expected []value.ScheduleId
	}{
		{
			Location: mustParseTimezone("+00:00"),
			Expected: []value.ScheduleId{1, 2, 4, 5},
		},
		{
			Location: mustParseTimezone("+10:00"), // 22:00
			Expected: []value.ScheduleId{1, 4, 5},
		},
		{
			Location: mustParseTimezone("-23:00"), // day before
			Expected: []value.ScheduleId{1, 2, 3, 4, 5},
		},
	}

	for i, testCase := range testCases {
		ctx := context.Background()
		now := time.Now().In(testCase.Location)

		ids := uc.getActualSchedulesIds(ctx, testSchedules, now, testCase.Location)

		require.Equalf(t, testCase.Expected, ids, "test case: %d", i+1)
	}

}

func TestGetSchedule(t *testing.T) {
	testSchedules := []entity.Schedule{
		testSchedules[0],
		testSchedules[1],
		testSchedules[3],
		testSchedules[4],
		{
			Id:     6,
			UserId: testUser,
			Name:   "Test Schedule 6",
			Period: value.SchedulePeriod(day),
		},
	}

	expected := []value.ScheduleTimeTable{
		{ // 0
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 8)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 9)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 10)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 11)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 12)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 13)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 14)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 16)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 17)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 18)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 19)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 20)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 21)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 22)),
		},
		{ // 1
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 8)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 20)),
		},
		{ // 3

			value.NewScheduleTimeTableItem(date().Add(time.Hour * 8)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 9)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 10)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 11)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*12 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*13 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*14 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*15 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*16 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*17 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*18 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*19 + time.Minute*15)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*20 + time.Minute*30)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour*21 + time.Minute*30)),
		},
		{ // 4
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 8)),
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 22)),
		},
		{
			value.NewScheduleTimeTableItem(date().Add(time.Hour * 8)),
		},
	}

	uc := NewUsecase(nil, l, testConfig)

	for i, testSchedule := range testSchedules {

		resp := uc.makeTimetable(context.Background(), &testSchedule, time.Now(), time.UTC)

		require.Equalf(t, expected[i], resp, "test case: %d", i+1)
	}
}

func TestGetNextTaking(t *testing.T) {
	uc := NewUsecase(nil, l, testConfig)

	testCases := []struct {
		Location         *time.Location
		NextTakingPeriod time.Duration
		Expected         []entity.ScheduleNextTaking
	}{
		{
			NextTakingPeriod: testConfig.NextTakingPeriod,
			Location:         mustParseTimezone("+00:00"), // 12:00
		},
		{
			NextTakingPeriod: testConfig.NextTakingPeriod,
			Location:         mustParseTimezone("-04:00"), // 08:00
		},
		{
			NextTakingPeriod: testConfig.NextTakingPeriod,
			Location:         mustParseTimezone("+09:00"), // 21:00
		},
		{
			NextTakingPeriod: testConfig.NextTakingPeriod,
			Location:         mustParseTimezone("+10:00"), // 22:00
		},
		{
			NextTakingPeriod: testConfig.NextTakingPeriod,
			Location:         mustParseTimezone("+19:30"), // 07:30 next day
		},
		{
			NextTakingPeriod: time.Hour * 13,              // until 9:00 next day
			Location:         mustParseTimezone("-16:00"), // 20:00 day before
		},
	}

	testCases[0].Expected = []entity.ScheduleNextTaking{
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[0].Location).Add(time.Hour*12 + time.Minute*15)),
		},
		{
			Id:         testSchedules[0].Id,
			Name:       testSchedules[0].Name,
			EndAt:      testSchedules[0].EndAt,
			Period:     testSchedules[0].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[0].Location).Add(time.Hour*13 + time.Minute*0)),
		},
	}
	testCases[1].Expected = []entity.ScheduleNextTaking{
		{
			Id:         testSchedules[0].Id,
			Name:       testSchedules[0].Name,
			EndAt:      testSchedules[0].EndAt,
			Period:     testSchedules[0].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[1].Location).Add(time.Hour*9 + time.Minute*0)),
		},
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[1].Location).Add(time.Hour*9 + time.Minute*0)),
		},
	}
	testCases[2].Expected = []entity.ScheduleNextTaking{
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[2].Location).Add(time.Hour*21 + time.Minute*30)),
		},
	}
	testCases[3].Expected = []entity.ScheduleNextTaking{}
	testCases[4].Expected = []entity.ScheduleNextTaking{
		{
			Id:         testSchedules[0].Id,
			Name:       testSchedules[0].Name,
			EndAt:      testSchedules[0].EndAt,
			Period:     testSchedules[0].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[4].Location).Add(time.Hour*8 + time.Minute*0 + day)),
		},
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[4].Location).Add(time.Hour*8 + time.Minute*0 + day)),
		},
		{
			Id:         testSchedules[4].Id,
			Name:       testSchedules[4].Name,
			EndAt:      testSchedules[4].EndAt,
			Period:     testSchedules[4].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[4].Location).Add(time.Hour*8 + time.Minute*0 + day)),
		},
	}
	testCases[5].Expected = []entity.ScheduleNextTaking{
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*20 + time.Minute*30 - day)),
		},
		{
			Id:         testSchedules[0].Id,
			Name:       testSchedules[0].Name,
			EndAt:      testSchedules[0].EndAt,
			Period:     testSchedules[0].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*21 + time.Minute*0 - day)),
		},
		{
			Id:         testSchedules[2].Id,
			Name:       testSchedules[2].Name,
			EndAt:      testSchedules[2].EndAt,
			Period:     testSchedules[2].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*21 + time.Minute*0 - day)),
		},
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*21 + time.Minute*30 - day)),
		},

		// next day
		{
			Id:         testSchedules[0].Id,
			Name:       testSchedules[0].Name,
			EndAt:      testSchedules[0].EndAt,
			Period:     testSchedules[0].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*8 + time.Minute*0)),
		},
		{
			Id:         testSchedules[1].Id,
			Name:       testSchedules[1].Name,
			EndAt:      testSchedules[1].EndAt,
			Period:     testSchedules[1].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*8 + time.Minute*0)),
		},
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*8 + time.Minute*0)),
		},
		{
			Id:         testSchedules[4].Id,
			Name:       testSchedules[4].Name,
			EndAt:      testSchedules[4].EndAt,
			Period:     testSchedules[4].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*8 + time.Minute*0)),
		},
		{
			Id:         testSchedules[0].Id,
			Name:       testSchedules[0].Name,
			EndAt:      testSchedules[0].EndAt,
			Period:     testSchedules[0].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*9 + time.Minute*0)),
		},
		{
			Id:         testSchedules[3].Id,
			Name:       testSchedules[3].Name,
			EndAt:      testSchedules[3].EndAt,
			Period:     testSchedules[3].Period,
			NextTaking: value.NewScheduleNextTaking(date(testCases[5].Location).Add(time.Hour*9 + time.Minute*0)),
		},
	}

	for i, c := range testCases {
		fmt.Printf("\nTast case %d\n\n", i+1)

		for j, s := range c.Expected {
			if !s.EndAt.IsNil() {
				c.Expected[j].EndAt = value.NewScheduleEndAt(util.Ptr(time.Date(s.EndAt.Year(), s.EndAt.Month(), s.EndAt.Day(), testConfig.EndDayHour, 0, 0, 0, c.Location)))
			}
		}

		uc.cfg.NextTakingPeriod = c.NextTakingPeriod

		ctx := context.Background()
		now := time.Now().In(c.Location)

		resp := uc.findNextTakings(ctx, testSchedules, now, c.Location)

		require.Equalf(t, c.Expected, resp, "test case: %d", i+1)
	}
}

func date(loc ...*time.Location) time.Time {
	usingLoc := time.UTC
	if len(loc) != 0 {
		if loc[0] != nil {
			usingLoc = loc[0]
		}
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, usingLoc)
}

func mustParseTimezone(name string) *time.Location {
	loc, err := util.ParseTimezone(name)
	if err != nil {
		panic(err)
	}
	return loc
}
