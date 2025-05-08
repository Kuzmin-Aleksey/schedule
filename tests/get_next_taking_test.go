package tests

import (
	"context"
	"errors"
	"net/http"
	"schedule/internal/util"
	"schedule/pkg/dbtest"
	schedulev1 "schedule/pkg/grpc"
	"schedule/pkg/rest"
	"time"
)

func (s *Suite) TestGetNextTakingHTTP() {
	const (
		userId = 1000000000000000
	)

	rq := s.Require()
	ctx := context.Background()

	err := dbtest.MigrateFromFile(s.db, "testdata/get_next_taking.sql")
	rq.NoError(err)

	testCases := []struct {
		name             string
		bootstrap        func()
		request          rest.GetNextTakingParams
		expectedResponse []rest.NextTakingResponse
		expectedStatus   int
		expectedError    rest.ErrorResponse
	}{
		{
			name: "success",
			request: rest.GetNextTakingParams{
				UserId: userId,
			},
			expectedResponse: []rest.NextTakingResponse{
				{
					Id:         2,
					Name:       "Test get_next_taking name2",
					EndAt:      nil,
					Period:     (time.Minute * 70).String(),
					NextTaking: time.Date(2025, time.January, 1, 12, 45, 0, 0, time.UTC).Format(time.RFC3339),
				},
				{
					Id:         1,
					Name:       "Test get_next_taking name1",
					EndAt:      util.Ptr(time.Date(2025, time.January, 1, s.cfg.Schedule.EndDayHour, 0, 0, 0, time.UTC).Format(time.RFC3339)),
					Period:     (time.Minute * 60).String(),
					NextTaking: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
				{
					Id:         3,
					Name:       "Test get_next_taking name3",
					EndAt:      util.Ptr(time.Date(2025, time.January, 2, s.cfg.Schedule.EndDayHour, 0, 0, 0, time.UTC).Format(time.RFC3339)),
					Period:     (time.Minute * 60 * 5).String(),
					NextTaking: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC).Format(time.RFC3339),
				},
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.httpClient.GetNextTakingWithResponse(ctx, &tc.request)
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

func (s *Suite) TestGetNextTakingGRPC() {
	const (
		userId = 1000000000000000
	)

	rq := s.Require()
	ctx := context.Background()

	err := dbtest.MigrateFromFile(s.db, "testdata/get_next_taking.sql")
	rq.NoError(err)

	testCases := []struct {
		name             string
		bootstrap        func()
		request          schedulev1.GetNextTakingsRequest
		expectedResponse schedulev1.GetNextTakingsReply
		expectedError    error
	}{
		{
			name: "success",
			request: schedulev1.GetNextTakingsRequest{
				UserId: userId,
			},
			expectedResponse: schedulev1.GetNextTakingsReply{
				Items: []*schedulev1.GetNextTakingsReplyItem{
					{
						Id:         2,
						Name:       "Test get_next_taking name2",
						EndAt:      0,
						Period:     int64(time.Minute * 70),
						NextTaking: time.Date(2025, time.January, 1, 12, 45, 0, 0, time.UTC).Unix(),
					},
					{
						Id:         1,
						Name:       "Test get_next_taking name1",
						EndAt:      time.Date(2025, time.January, 1, s.cfg.Schedule.EndDayHour, 0, 0, 0, time.UTC).Unix(),
						Period:     int64(time.Minute * 60),
						NextTaking: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC).Unix(),
					},
					{
						Id:         3,
						Name:       "Test get_next_taking name3",
						EndAt:      time.Date(2025, time.January, 2, s.cfg.Schedule.EndDayHour, 0, 0, 0, time.UTC).Unix(),
						Period:     int64(time.Minute * 60 * 5),
						NextTaking: time.Date(2025, time.January, 1, 13, 0, 0, 0, time.UTC).Unix(),
					},
				},
			},
		},
	}

	for _, tc := range testCases { //nolint:govet
		s.Run(tc.name, func() {
			if tc.bootstrap != nil {
				tc.bootstrap()
			}

			resp, err := s.grpcClient.GetNextTakings(ctx, &tc.request)
			if tc.expectedError != nil {
				rq.Equal(tc.expectedError, err)
				return
			}
			rq.NoError(err)

			rq.Equal(len(tc.expectedResponse.GetItems()), len(resp.GetItems()))

			for i, respItem := range resp.GetItems() {
				expectedItem := tc.expectedResponse.GetItems()[i]

				rq.Equal(expectedItem.Id, respItem.Id)
				rq.Equal(expectedItem.Name, respItem.Name)
				rq.Equal(expectedItem.EndAt, respItem.EndAt)
				rq.Equal(expectedItem.Period, respItem.Period)
				rq.Equal(expectedItem.NextTaking, respItem.NextTaking)
			}

		})
	}
}
