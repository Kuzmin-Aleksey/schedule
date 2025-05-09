package tests

import (
	"context"
	"errors"
	"net/http"
	"schedule/internal/util"
	"schedule/pkg/dbtest"
	"schedule/pkg/errcodes"
	schedulev1 "schedule/pkg/grpc"
	"schedule/pkg/rest"
	"time"
)

func (s *Suite) TestGetScheduleHTTP() {
	const (
		userId     = 1000000000000000
		scheduleId = 1
	)

	rq := s.Require()
	ctx := context.Background()

	err := dbtest.MigrateFromFile(s.db, "testdata/get_schedule.sql")
	rq.NoError(err)

	testCases := []struct {
		name             string
		bootstrap        func()
		request          rest.GetScheduleParams
		expectedResponse rest.ScheduleResponse
		expectedStatus   int
		expectedError    rest.ErrorResponse
	}{
		{
			name: "success",
			request: rest.GetScheduleParams{
				UserId:     userId,
				ScheduleId: scheduleId,
			},
			expectedResponse: rest.ScheduleResponse{
				Id:     scheduleId,
				Name:   "Test get_schedule name",
				EndAt:  util.Ptr(time.Date(2025, time.January, 1, s.cfg.Schedule.EndDayHour, 0, 0, 0, time.UTC).Format(time.RFC3339)),
				Period: (time.Minute * 120).String(),
				Timetable: []string{
					"08:00:00",
					"10:00:00",
					"12:00:00",
					"14:00:00",
					"16:00:00",
					"18:00:00",
					"20:00:00",
					"22:00:00",
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "not found",
			request: rest.GetScheduleParams{
				UserId:     userId,
				ScheduleId: -1,
			},
			expectedStatus: http.StatusNotFound,
			expectedError: rest.ErrorResponse{
				Error: errcodes.NotFound.String(),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.httpClient.GetScheduleWithResponse(ctx, &tc.request)
			rq.NoError(err)

			statusCode := resp.StatusCode()

			rq.Equal(tc.expectedStatus, statusCode)

			switch statusCode {
			case http.StatusOK:
				rq.EqualValues(&tc.expectedResponse, resp.JSON200)
			case http.StatusBadRequest:
				rq.Equal(&tc.expectedError, resp.JSON400)
			case http.StatusNotFound:
				rq.Equal(&tc.expectedError, resp.JSON404)
			case http.StatusInternalServerError:
				rq.Equal(&tc.expectedError, resp.JSON500)
			default:
				rq.Errorf(errors.New("unexpected status code"), "Code: %d\n body: %s", statusCode, string(resp.Body))
			}
		})
	}
}

func (s *Suite) TestGetScheduleGRPC() {
	const (
		userId     = 1000000000000000
		scheduleId = 1
	)

	rq := s.Require()
	ctx := context.Background()

	err := dbtest.MigrateFromFile(s.db, "testdata/get_schedule.sql")
	rq.NoError(err)

	testCases := []struct {
		name             string
		bootstrap        func()
		request          schedulev1.GetScheduleRequest
		expectedResponse schedulev1.GetScheduleReply
		expectedError    error
	}{
		{
			name: "success",
			request: schedulev1.GetScheduleRequest{
				ScheduleId: scheduleId,
				UserId:     userId,
			},
			expectedResponse: schedulev1.GetScheduleReply{
				Name:   "Test get_schedule name",
				EndAt:  time.Date(2025, time.January, 1, s.cfg.Schedule.EndDayHour, 0, 0, 0, time.UTC).Unix(),
				Period: int64(time.Minute * 120),
				Timetable: []int64{
					time.Date(2025, time.January, 1, 8, 0, 0, 0, time.UTC).Unix(),
					time.Date(2025, time.January, 1, 10, 0, 0, 0, time.UTC).Unix(),
					time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC).Unix(),
					time.Date(2025, time.January, 1, 14, 0, 0, 0, time.UTC).Unix(),
					time.Date(2025, time.January, 1, 16, 0, 0, 0, time.UTC).Unix(),
					time.Date(2025, time.January, 1, 18, 0, 0, 0, time.UTC).Unix(),
					time.Date(2025, time.January, 1, 20, 0, 0, 0, time.UTC).Unix(),
					time.Date(2025, time.January, 1, 22, 0, 0, 0, time.UTC).Unix(),
				},
			},
		},
	}

	for _, tc := range testCases { //nolint:govet
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.grpcClient.GetSchedule(ctx, &tc.request)
			if tc.expectedError != nil {
				rq.Equal(tc.expectedError, err)
				return
			}
			rq.NoError(err)

			rq.Equal(tc.expectedResponse.GetName(), resp.GetName())
			rq.Equal(tc.expectedResponse.GetEndAt(), resp.GetEndAt())
			rq.Equal(tc.expectedResponse.GetPeriod(), resp.GetPeriod())
			rq.Equal(tc.expectedResponse.GetTimetable(), resp.GetTimetable())
		})
	}
}
