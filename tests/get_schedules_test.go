package tests

import (
	"context"
	"errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net/http"
	"schedule/pkg/dbtest"
	schedulev1 "schedule/pkg/grpc"
	"schedule/pkg/rest"
)

func (s *Suite) TestGetSchedulesHTTP() {
	const (
		userId = 1000000000000000
	)

	rq := s.Require()
	ctx := context.Background()

	err := dbtest.MigrateFromFile(s.db, "testdata/get_schedules.sql")
	rq.NoError(err)

	testCases := []struct {
		name             string
		bootstrap        func()
		request          rest.GetSchedulesParams
		expectedResponse []int
		expectedStatus   int
		expectedError    rest.ErrorResponse
	}{
		{
			name: "success",
			request: rest.GetSchedulesParams{
				UserId: userId,
			},
			expectedResponse: []int{1, 2, 3},
			expectedStatus:   http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.httpClient.GetSchedulesWithResponse(ctx, &tc.request)
			rq.NoError(err)

			statusCode := resp.StatusCode()

			rq.Equal(tc.expectedStatus, statusCode)

			switch statusCode {
			case http.StatusOK:
				rq.EqualValues(tc.expectedResponse, *resp.JSON200)
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

func (s *Suite) TestGetSchedulesGRPC() {
	const (
		userId = 1000000000000000
	)

	rq := s.Require()
	ctx := context.Background()

	err := dbtest.MigrateFromFile(s.db, "testdata/get_schedules.sql")
	rq.NoError(err)

	testCases := []struct {
		name             string
		bootstrap        func()
		request          schedulev1.GetSchedulesRequest
		expectedResponse schedulev1.GetSchedulesReply
		expectedCode     codes.Code
	}{
		{
			name: "success",
			request: schedulev1.GetSchedulesRequest{
				UserId: userId,
			},
			expectedResponse: schedulev1.GetSchedulesReply{
				ScheduleIds: []int32{1, 2, 3},
			},
		},
	}

	for _, tc := range testCases { //nolint:govet
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.grpcClient.GetSchedules(ctx, &tc.request)

			statusCode := status.Code(err)
			rq.Equal(tc.expectedCode, statusCode)

			if statusCode != codes.OK {
				return
			}

			rq.NoError(err)

			rq.Equal(tc.expectedResponse.GetScheduleIds(), resp.GetScheduleIds())
		})
	}
}
