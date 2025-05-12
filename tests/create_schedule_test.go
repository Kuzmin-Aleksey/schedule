package tests

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"schedule/internal/domain/entity"
	"schedule/internal/domain/value"
	"schedule/internal/util"
	schedulev1 "schedule/pkg/grpc"
	"schedule/pkg/rest"
	"time"
)

func (s *Suite) TestCreateScheduleHTTP() {
	const (
		userId = 1000000000000000
	)

	rq := s.Require()
	ctx := context.Background()

	testCases := []struct {
		name           string
		bootstrap      func()
		request        rest.CreateScheduleRequest
		expectedStatus int
		expectedError  rest.ErrorResponse
		expectedData   entity.Schedule
	}{
		{
			name: "success",
			request: rest.CreateScheduleRequest{
				UserId:   userId,
				Name:     "Test name",
				Period:   time.Hour.String(),
				Duration: 10,
			},
			expectedStatus: http.StatusOK,
			expectedData: entity.Schedule{
				UserId: userId,
				Name:   "Test name",
				Period: value.SchedulePeriod(time.Hour),
				EndAt:  value.NewScheduleEndAt(util.Ptr(time.Date(2025, time.January, 11, 0, 0, 0, 0, time.UTC))),
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.httpClient.PostScheduleWithResponse(ctx, tc.request)
			rq.NoError(err)

			statusCode := resp.StatusCode()

			rq.Equal(tc.expectedStatus, statusCode)

			switch statusCode {
			case http.StatusOK:
				tc.expectedData.Id = value.ScheduleId(resp.JSON200.Id)

				var data entity.Schedule

				err = s.db.GetContext(ctx, &data, "SELECT * FROM schedule WHERE id = ?", resp.JSON200.Id)
				rq.NoError(err)

				rq.Equal(tc.expectedData, data)

			case http.StatusBadRequest:
				rq.Equal(tc.expectedError, resp.JSON400)
			case http.StatusInternalServerError:
				rq.Equal(tc.expectedError, resp.JSON500)
			default:
				rq.Errorf(errors.New("unexpected status code"), "Code: %d\n body: %s", statusCode, string(resp.Body))
			}
		})
	}
}

func (s *Suite) TestCreateScheduleGRPC() {
	const (
		userId = 1000000000000000
	)

	rq := s.Require()
	ctx := context.Background()

	testCases := []struct {
		name             string
		bootstrap        func()
		request          schedulev1.CreateScheduleRequest
		expectedResponse schedulev1.CreateScheduleReply
		expectedCode     codes.Code
		expectedData     entity.Schedule
	}{
		{
			name: "success",
			request: schedulev1.CreateScheduleRequest{
				UserId:   userId,
				Name:     "Test name",
				Period:   int64(time.Hour),
				Duration: 10,
			},
			expectedData: entity.Schedule{
				UserId: userId,
				Name:   "Test name",
				Period: value.SchedulePeriod(time.Hour),
				EndAt:  value.NewScheduleEndAt(util.Ptr(time.Date(2025, time.January, 11, 0, 0, 0, 0, time.UTC))),
			},
		},
	}

	for _, tc := range testCases { //nolint:govet
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.grpcClient.CreateSchedule(ctx, &tc.request)

			statusCode := status.Code(err)
			rq.Equal(tc.expectedCode, statusCode)

			if statusCode != codes.OK {
				return
			}

			rq.NoError(err)

			tc.expectedData.Id = value.ScheduleId(resp.GetId())

			var data entity.Schedule

			err = s.db.GetContext(ctx, &data, "SELECT * FROM schedule WHERE id = ?", resp.GetId())
			rq.NoError(err)

			rq.Equal(tc.expectedData, data)
		})
	}
}
