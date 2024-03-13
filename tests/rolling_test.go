package tests

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_Maintenance"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Auth"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Cms"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Discovery"
	"github.com/ydb-platform/ydbops/pkg/options"
	blackmagic "github.com/ydb-platform/ydbops/tests/black-magic"
	"github.com/ydb-platform/ydbops/tests/mock"
	"google.golang.org/protobuf/proto"
)

func prepareEnvVariables() map[string]string {
	previous := make(map[string]string)

	newValue := mock.TestPassword
	os.Setenv(options.DefaultStaticPasswordEnvVar, newValue)
	previous[options.DefaultStaticPasswordEnvVar] = os.Getenv(options.DefaultStaticPasswordEnvVar)

	return previous
}

func revertEnvVariables(previous map[string]string) {
	for k, v := range previous {
		os.Setenv(k, v)
	}
}

var _ = Describe("Test Rolling", func() {
	var ydb *mock.YdbMock
	var previousEnvVars map[string]string

	BeforeEach(func() {
		port := 2135
		ydb = mock.NewYdbMockServer()
		ydb.StartOn(port)
		previousEnvVars = prepareEnvVariables()
	})

	AfterEach(func() {
		ydb.Teardown()
		revertEnvVariables(previousEnvVars)
	})

	It("happy path: restart 3 out of 8 nodes, strong mode, no failures", func() {

		cmd := exec.Command(filepath.Join("..", "ydbops"),
			"--endpoint", "grpc://localhost:2135",
			"--verbose",
			"restart",
			"--availability-mode", "strong",
			"--hosts=1,2,3",
			"--user", mock.TestUser,
			"run",
			"--payload", filepath.Join(".", "mock", "noop-payload.sh"),
		)

		_, err := cmd.CombinedOutput()
		// fmt.Println(string(output))

		Expect(err).To(BeNil())

		if err != nil {
			fmt.Println("Error getting combined output:", err)
			return
		}

		expectedRequests := []proto.Message{
			&Ydb_Auth.LoginRequest{
				User:     mock.TestUser,
				Password: mock.TestPassword,
			},
			&Ydb_Cms.ListDatabasesRequest{},
			&Ydb_Maintenance.ListClusterNodesRequest{},
			&Ydb_Discovery.WhoAmIRequest{},
			&Ydb_Maintenance.ListMaintenanceTasksRequest{
				User: &mock.TestUser,
			},
			&Ydb_Maintenance.CreateMaintenanceTaskRequest{
				TaskOptions: &Ydb_Maintenance.MaintenanceTaskOptions{
					TaskUid:          "task-UUID-1",
					Description:      "Rolling restart maintenance task",
					AvailabilityMode: Ydb_Maintenance.AvailabilityMode_AVAILABILITY_MODE_STRONG,
				},
				ActionGroups: []*Ydb_Maintenance.ActionGroup{
					{}, {}, {},
				},
			},
			&Ydb_Maintenance.CompleteActionRequest{
				ActionUids: []*Ydb_Maintenance.ActionUid{
					{
						TaskUid:  "task-UUID-1",
						GroupId:  "group-UUID-1",
						ActionId: "action-UUID-1",
					},
				},
			},
			&Ydb_Maintenance.RefreshMaintenanceTaskRequest{
				TaskUid: "task-UUID-1",
			},
			&Ydb_Maintenance.CompleteActionRequest{
				ActionUids: []*Ydb_Maintenance.ActionUid{
					{
						TaskUid:  "task-UUID-1",
						GroupId:  "group-UUID-2",
						ActionId: "action-UUID-2",
					},
				},
			},
			&Ydb_Maintenance.RefreshMaintenanceTaskRequest{
				TaskUid: "task-UUID-1",
			},
			&Ydb_Maintenance.CompleteActionRequest{
				ActionUids: []*Ydb_Maintenance.ActionUid{
					{
						TaskUid:  "task-UUID-1",
						GroupId:  "group-UUID-3",
						ActionId: "action-UUID-3",
					},
				},
			},
		}

		actualRequests := ydb.RequestLog

		// for _, req := range actualRequests {
		// 	fmt.Printf("\n%+v : %+v\n", reflect.TypeOf(req), req)
		// }

		Expect(len(expectedRequests)).To(Equal(len(actualRequests)))

		values := make(map[string]string)
		for i, expected := range expectedRequests {
			actual := actualRequests[i]
			blackmagic.DeepEqualOnPresentFields(expected, actual, values)
		}
	})
})