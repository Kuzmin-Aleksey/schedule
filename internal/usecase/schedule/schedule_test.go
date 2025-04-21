package schedule

import (
	"context"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"schedule/config"
	"schedule/internal/entity"
	"schedule/internal/usecase/schedule/mocks"
	"schedule/internal/util"
	"testing"
	"time"
)

func init() {
	time.Local = nil
}

var testConfig = config.ScheduleConfig{
	NextTakingPeriod: time.Hour,
	BeginDayHour:     8,
	EndDayHour:       22,
	TimeRound:        time.Minute * 15,
}

const testUser int64 = 1234567890123456

var testSchedules = []entity.Schedule{
	{
		Id:     1,
		UserId: testUser,
		Name:   "Test Schedule 1",
		EndAt:  util.Ptr(date().Add(day)),
		Period: time.Hour,
	},
	{
		Id:     2,
		UserId: testUser,
		Name:   "Test Schedule 2",
		EndAt:  util.Ptr(date()),
		Period: time.Hour * 12,
	},
	{
		Id:     3,
		UserId: testUser,
		Name:   "Test Schedule 3",
		EndAt:  util.Ptr(date().Add(-day)),
		Period: time.Hour * 5,
	},
	{
		Id:     4,
		UserId: testUser,
		Name:   "Test Schedule 3",
		Period: time.Hour + time.Minute*2,
	},
}

func TestGetByUser(t *testing.T) {
	expected := []int{1, 2, 4}

	r := mocks.NewRepo(t)
	r.On("GetByUser", mock.Anything, testUser).Return(testSchedules, nil)

	uc := NewUsecase(r, testConfig)

	ctx := context.Background()
	if time.Now().Hour() >= 22 {
		loc, _ := util.ParseTimezone("-02:00")
		ctx = CtxWithLocation(ctx, loc)
	}

	ids, err := uc.GetByUser(ctx, testUser)
	require.NoError(t, err)

	require.Equal(t, expected, ids)
}

func TestGetSchedule(t *testing.T) {
	expected := []ScheduleResponseDTO{
		{
			Id:     testSchedules[0].Id,
			Name:   testSchedules[0].Name,
			EndAt:  testSchedules[0].EndAt,
			Period: util.JsonDuration(testSchedules[0].Period),
			Timetable: []time.Time{
				date().Add(time.Hour * 8),
				date().Add(time.Hour * 9),
				date().Add(time.Hour * 10),
				date().Add(time.Hour * 11),
				date().Add(time.Hour * 12),
				date().Add(time.Hour * 13),
				date().Add(time.Hour * 14),
				date().Add(time.Hour * 15),
				date().Add(time.Hour * 16),
				date().Add(time.Hour * 17),
				date().Add(time.Hour * 18),
				date().Add(time.Hour * 19),
				date().Add(time.Hour * 20),
				date().Add(time.Hour * 21),
				date().Add(time.Hour * 22),
			},
		},
		{
			Id:     testSchedules[1].Id,
			Name:   testSchedules[1].Name,
			EndAt:  testSchedules[1].EndAt,
			Period: util.JsonDuration(testSchedules[1].Period),
			Timetable: []time.Time{
				date().Add(time.Hour * 8),
				date().Add(time.Hour * 20),
			},
		},
		{
			Id:        testSchedules[2].Id,
			Name:      testSchedules[2].Name,
			EndAt:     testSchedules[2].EndAt,
			Period:    util.JsonDuration(testSchedules[2].Period),
			Timetable: []time.Time{},
		},
		{
			Id:     testSchedules[3].Id,
			Name:   testSchedules[3].Name,
			EndAt:  testSchedules[3].EndAt,
			Period: util.JsonDuration(testSchedules[3].Period),
			Timetable: []time.Time{
				date().Add(time.Hour * 8),
				date().Add(time.Hour * 9),
				date().Add(time.Hour * 10),
				date().Add(time.Hour * 11),
				date().Add(time.Hour*12 + time.Minute*15),
				date().Add(time.Hour*13 + time.Minute*15),
				date().Add(time.Hour*14 + time.Minute*15),
				date().Add(time.Hour*15 + time.Minute*15),
				date().Add(time.Hour*16 + time.Minute*15),
				date().Add(time.Hour*17 + time.Minute*15),
				date().Add(time.Hour*18 + time.Minute*15),
				date().Add(time.Hour*19 + time.Minute*15),
				date().Add(time.Hour*20 + time.Minute*30),
				date().Add(time.Hour*21 + time.Minute*30),
			},
		},
	}
	b, _ := expected[0].Timetable[0].MarshalJSON()
	t.Log(string(b), "-----------")

	r := mocks.NewRepo(t)
	uc := NewUsecase(r, testConfig)

	for i, testSchedule := range testSchedules {
		r.On("GetById", mock.Anything, testUser, testSchedule.Id).Return(&testSchedule, nil)

		resp, err := uc.GetTimetable(context.Background(), testSchedule.UserId, testSchedule.Id)
		require.NoError(t, err)
		require.Equal(t, &expected[i], resp)
	}

}

func TestGetNextTaking(t *testing.T) {
	r := mocks.NewRepo(t)
	r.On("GetByUser", mock.Anything, testUser).Return(testSchedules, nil)

	uc := NewUsecase(r, testConfig)

	tz, _ := time.Parse("-07:00", "-04:00")

	resp, err := uc.GetNextTakings(context.WithValue(context.Background(), userLocationCtxKey{}, tz.Location()), testUser)
	require.NoError(t, err)
	require.NotNil(t, resp)

	now := time.Now().In(tz.Location())
	t.Log("user time:", now)

	for _, nt := range resp {
		t.Logf("schedule: %s, next taking: %s", nt.Name, nt.NextTaking.String())
	}
}

func date() time.Time {
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
}
