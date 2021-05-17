package service_test

import (
	"context"
	_ "google.golang.org/genproto/protobuf/field_mask"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"math/rand"
	"testing"

	apiProto "github.com/anchamber/genetics-api/proto"
	"github.com/anchamber/genetics-tank/db"
	sm "github.com/anchamber/genetics-tank/db/model"
	tankProto "github.com/anchamber/genetics-tank/proto"
	"github.com/anchamber/genetics-tank/service"
	grpc "google.golang.org/grpc"
)

var testData = []*sm.Tank{
	}

var testTanksToCreate = []*sm.Tank{
	}

func TestStreamTanks(t *testing.T) {
	testCases := []struct {
		name          string
		request       *tankProto.StreamTanksRequest
		responses     []*sm.Tank
		expectedError bool
	}{
		{
			name:          "request all entries",
			request:       &tankProto.StreamTanksRequest{},
			responses:     db.MockDataTanks,
			expectedError: false,
		},
		{
			name: "request with limit",
			request: &tankProto.StreamTanksRequest{
				Pageination: &apiProto.Pagination{
					Limit: 2,
				},
			},
			responses:     db.MockDataTanks[0:2],
			expectedError: false,
		},
		{
			name: "request with offset",
			request: &tankProto.StreamTanksRequest{
				Pageination: &apiProto.Pagination{
					Offset: 2,
				},
			},
			responses:     db.MockDataTanks[2:],
			expectedError: false,
		},
		{
			name: "request with offset and limit",
			request: &tankProto.StreamTanksRequest{
				Pageination: &apiProto.Pagination{
					Offset: 2,
					Limit:  1,
				},
			},
			responses:     db.MockDataTanks[2:3],
			expectedError: false,
		},
		{
			name: "request with name filter EQ",
			request: &tankProto.StreamTanksRequest{
				Filters: []*apiProto.Filter{
					{
						Key:      "name",
						Operator: apiProto.Operator_EQ,
					},
				},
			},
			responses: []*sm.Tank{
				db.MockDataTanks[1],
			},
			expectedError: false,
		},
		{
			name: "request with name filter CONTAINS",
			request: &tankProto.StreamTanksRequest{
				Filters: []*apiProto.Filter{
					{
						Key:      "name",
						Operator: apiProto.Operator_CONTAINS,
					},
				},
			},
			responses: []*sm.Tank{
				db.MockDataTanks[2],
			},
			expectedError: false,
		},
		{
			name: "request with invalid filter key",
			request: &tankProto.StreamTanksRequest{
				Filters: []*apiProto.Filter{
					{
						Key:      "nme",
						Operator: apiProto.Operator_CONTAINS,
					},
				},
			},
			responses:     db.MockDataTanks,
			expectedError: false,
		},
	}

	tankServer := service.New(db.NewMockDB(testData))
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			t.Log(tc.responses)
			serviceMock := MockTankService{
				t:         t,
				responses: tc.responses,
			}
			err := tankServer.StreamTanks(tc.request, &serviceMock)
			if validateError(t, err, codes.Unknown, tc.expectedError) {
				return
			}
			if serviceMock.CallCount != len(tc.responses) {
				t.Errorf("Call count of mock does not match, expected: %d | actual: %d", len(tc.responses), serviceMock.CallCount)
			}
		})
	}
}

func TestGetTank(t *testing.T) {
	index := rand.Intn(len(testData))
	testCases := []struct {
		name          string
		request       *tankProto.GetTankRequest
		response      *sm.Tank
		expectedError bool
		errorCode     codes.Code
	}{
		{
			name:          "request existing tank",
			response:      testData[index],
			expectedError: false,
			request: &tankProto.GetTankRequest{
			},
		},
		{
			name:          "request none existing tank",
			response:      nil,
			expectedError: true,
			request: &tankProto.GetTankRequest{
				Number: 0,
			},
			errorCode: codes.NotFound,
		},
	}

	tankServer := service.New(db.NewMockDB(testData))
	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			resp, err := tankServer.GetTank(context.Background(), tc.request)
			if validateError(t, err, tc.errorCode, tc.expectedError) {
				return
			}
			compareResponseToTank(t, resp, tc.response)
		})
	}
}

func TestCreateTank(t *testing.T) {
	tank := testTanksToCreate[rand.Intn(len(testTanksToCreate))]
	testCases := []struct {
		name          string
		request       *tankProto.CreateTankRequest
		response      *sm.Tank
		expectedError bool
		errorCode     codes.Code
	}{
		{
			name:          "create valid tank",
			response:      tank,
			expectedError: false,
			request: &tankProto.CreateTankRequest{
			},
		},
		{
			name:          "create tank with invalid name",
			response:      nil,
			expectedError: true,
			request: &tankProto.CreateTankRequest{
			},
			errorCode: codes.InvalidArgument,
		},
		{
			name:          "create tank with invalid cleaning interval",
			response:      nil,
			expectedError: true,
			request: &tankProto.CreateTankRequest{
			},
			errorCode: codes.InvalidArgument,
		},
	}

	for _, tc := range testCases {
		tankServer := service.New(db.NewMockDB(testData))
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			res, err := tankServer.CreateTank(context.Background(), tc.request)
			if validateError(t, err, tc.errorCode, tc.expectedError) {
				return
			}
			if res == nil {
				t.Error("response should not be nil")
			}
		})
	}
}

func TestUpdateTank(t *testing.T) {
	tank := testData[rand.Intn(len(testData))]
	testCases := []struct {
		name          string
		request       *tankProto.UpdateTankRequest
		expected      sm.Tank
		expectedError bool
		errorCode     codes.Code
	}{
	}

	for _, tc := range testCases {
		tankServer := service.New(db.NewMockDB(testData))
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := tankServer.UpdateTank(context.Background(), tc.request)
			validateError(t, err, tc.errorCode, tc.expectedError)
			resp, err := tankServer.GetTank(context.Background(), &tankProto.GetTankRequest{Name: tc.expected.Name})
			if validateError(t, err, tc.errorCode, tc.expectedError) {
				return
			}
			compareResponseToTank(t, resp, &tc.expected)
		})
	}
}

func TestDeleteTank(t *testing.T) {
	tank := testData[rand.Intn(len(testData))]
	testCases := []struct {
		name             string
		request          *tankProto.DeleteTankRequest
		expectedErrorGet bool
		expectedErrorDel bool
		errorCode        codes.Code
	}{
		{
			name: "delete existing tank",
			request: &tankProto.DeleteTankRequest{
			},
			expectedErrorDel: false,
			expectedErrorGet: true,
			errorCode:        codes.NotFound,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tankServer := service.New(db.NewMockDB(testData))
			_, err := tankServer.DeleteTank(context.Background(), tc.request)
			if validateError(t, err, tc.errorCode, tc.expectedErrorDel) {
				return
			}
			resp, err := tankServer.GetTank(context.Background(), &tankProto.GetTankRequest{Name: tank.Number})
			if validateError(t, err, tc.errorCode, tc.expectedErrorGet) {
				return
			}
			compareResponseToTank(t, resp, tank)
		})
	}
}

type MockTankService struct {
	CallCount int
	t         *testing.T
	responses []*sm.Tank
	grpc.ServerStream
}

func (x *MockTankService) Send(resp *tankProto.TankResponse) error {
	compareResponseToTank(x.t, resp, x.responses[x.CallCount])
	x.CallCount++
	return nil
}

func compareResponseToTank(t *testing.T, resp *tankProto.TankResponse, tank *sm.Tank) {

	if tank.Number != resp.Number {
		t.Errorf("numbers do not match, expected: %d | actual: %d", tank.Number, resp.Number)
	}
	if tank.System != resp.System {
		t.Errorf("systems do not match, expected: %s | actual: %s", tank.System, resp.System)
	}
	if tank.Active != resp.Active {
		t.Errorf("locations do not match, expected: %v | actual: %v", tank.Active, resp.Active)
	}
	if tank.Size != resp.Size {
		t.Errorf("cleaning intervals do not match, expected: %d | actual: %d", tank.Size, resp.Size)
	}
	if tank.FishCount != resp.FishCount {
		t.Errorf("last cleaned do not match, expected: %d | actual: %d", tank.FishCount, resp.FishCount)
	}
}

func validateError(t *testing.T, err error, code codes.Code, expected bool) bool {
	done := false
	if err != nil {
		if !expected {
			t.Fatalf("response returned error and when it should be ok: %v", err)
		}
		st, ok := status.FromError(err)
		if !ok {
			t.Fatalf("was not a status: %v", err)
		}
		if st.Code() != code {
			t.Fatalf("wrong status code: expected %v | actual: %v", code, st.Code())
		}
		done = true
	} else {
		if expected {
			t.Fatalf("response should have had an error")
		}
	}
	return done
}
