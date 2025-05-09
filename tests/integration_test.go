package tests

import (
	"bou.ke/monkey"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/suite"
	"golang.org/x/sys/windows"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"os"
	"runtime"
	"schedule/internal/app"
	"schedule/internal/config"
	"schedule/internal/infrastructure/persistence/mysql"
	"schedule/pkg/dbtest"
	schedulev1 "schedule/pkg/grpc"
	"schedule/pkg/rest"
	"sync"
	"testing"
	"time"
)

func init() {
	now := time.Date(2025, time.January, 1, 12, 0, 0, 0, time.UTC)
	monkey.Patch(time.Now, func() time.Time { return now })
	time.Local = nil
}

type Suite struct {
	suite.Suite

	wg  sync.WaitGroup
	cfg *config.Config
	db  *sqlx.DB

	httpClient rest.ClientWithResponsesInterface
	grpcClient schedulev1.ScheduleClient
}

func TestIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	suite.Run(t, &Suite{})
}

func (s *Suite) SetupSuite() {
	var err error

	rq := s.Require()

	s.cfg, err = config.ReadConfig("../config/config.yaml", "../.env")
	rq.NoError(err)

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		app.Run(s.cfg)
	}()

	// Wait for the application to start.
	time.Sleep(time.Second)

	s.db, err = mysql.Connect(s.cfg.MySQl)
	rq.NoError(err)

	s.httpClient, err = rest.NewClientWithResponses("http://" + s.cfg.HttpServer.Addr)
	rq.NoError(err)

	grpcConn, err := grpc.NewClient(s.cfg.GrpcServer.Addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	rq.NoError(err)

	s.grpcClient = schedulev1.NewScheduleClient(grpcConn)
}

func (s *Suite) SetupTest() {
	rq := s.Require()

	err := dbtest.MigrateFromFile(s.db, "testdata/cleanup.sql")
	rq.NoError(err)
}

func (s *Suite) TearDownSuite() {
	rq := s.Require()

	if runtime.GOOS == "windows" {
		rq.NoError(windows.GenerateConsoleCtrlEvent(windows.CTRL_C_EVENT, 0)) // not work
	} else {
		p, err := os.FindProcess(os.Getpid())
		rq.NoError(err)
		rq.NoError(p.Signal(os.Interrupt))
	}

	s.wg.Wait()

	rq.NoError(s.db.Close())

}
