package client

import (
	"context"
	"fmt"

	"github.com/ydb-platform/ydb-go-genproto/Ydb_Cms_V1"
	"github.com/ydb-platform/ydb-go-genproto/draft/Ydb_Maintenance_V1"
	"github.com/ydb-platform/ydb-go-genproto/draft/protos/Ydb_Maintenance"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Cms"
	"github.com/ydb-platform/ydb-go-genproto/protos/Ydb_Operations"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"

	"github.com/ydb-platform/ydbops/internal/collections"
)

type Cms struct {
	logger *zap.SugaredLogger
	f      *Factory
}

func NewCMSClient(f *Factory, logger *zap.SugaredLogger) *Cms {
	return &Cms{
		logger: logger,
		f:      f,
	}
}

func (c *Cms) Tenants() ([]string, error) {
	result := Ydb_Cms.ListDatabasesResult{}
	_, err := c.ExecuteCMSMethod(&result, func(ctx context.Context, cl Ydb_Cms_V1.CmsServiceClient) (OperationResponse, error) {
		c.logger.Debug("Invoke ListDatabases method")
		return cl.ListDatabases(ctx, &Ydb_Cms.ListDatabasesRequest{OperationParams: c.f.OperationParams()})
	})
	if err != nil {
		return nil, err
	}

	s := collections.SortBy(result.Paths,
		func(l string, r string) bool {
			return l < r
		},
	)
	return s, nil
}

func (c *Cms) Nodes() ([]*Ydb_Maintenance.Node, error) {
	result := Ydb_Maintenance.ListClusterNodesResult{}
	_, err := c.ExecuteMaintenanceMethod(&result,
		func(ctx context.Context, cl Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error) {
			c.logger.Debug("Invoke ListClusterNodes method")
			return cl.ListClusterNodes(ctx, &Ydb_Maintenance.ListClusterNodesRequest{OperationParams: c.f.OperationParams()})
		},
	)
	if err != nil {
		return nil, err
	}

	nodes := collections.SortBy(result.Nodes,
		func(l *Ydb_Maintenance.Node, r *Ydb_Maintenance.Node) bool {
			return l.NodeId < r.NodeId
		},
	)

	return nodes, nil
}

func (c *Cms) MaintenanceTasks(userSID string) ([]string, error) {
	result := Ydb_Maintenance.ListMaintenanceTasksResult{}
	_, err := c.ExecuteMaintenanceMethod(&result,
		func(ctx context.Context, cl Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error) {
			c.logger.Debug("Invoke ListMaintenanceTasks method")
			return cl.ListMaintenanceTasks(ctx,
				&Ydb_Maintenance.ListMaintenanceTasksRequest{
					OperationParams: c.f.OperationParams(),
					User:            &userSID,
				},
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return result.TasksUids, nil
}

func (c *Cms) GetMaintenanceTask(taskID string) (MaintenanceTask, error) {
	result := Ydb_Maintenance.GetMaintenanceTaskResult{}
	_, err := c.ExecuteMaintenanceMethod(&result,
		func(ctx context.Context, cl Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error) {
			c.logger.Debug("Invoke GetMaintenanceTask method")
			return cl.GetMaintenanceTask(ctx, &Ydb_Maintenance.GetMaintenanceTaskRequest{
				OperationParams: c.f.OperationParams(),
				TaskUid:         taskID,
			})
		},
	)
	if err != nil {
		return nil, err
	}

	return &maintenanceTaskResult{
		TaskUID:           taskID,
		ActionGroupStates: result.ActionGroupStates,
	}, nil
}

func (c *Cms) CreateMaintenanceTask(params MaintenanceTaskParams) (MaintenanceTask, error) {
	request := &Ydb_Maintenance.CreateMaintenanceTaskRequest{
		OperationParams: c.f.OperationParams(),
		TaskOptions: &Ydb_Maintenance.MaintenanceTaskOptions{
			TaskUid:          params.TaskUID,
			AvailabilityMode: params.AvailabilityMode,
			Description:      "Rolling restart maintenance task",
		},
		ActionGroups: make([]*Ydb_Maintenance.ActionGroup, 0, len(params.Nodes)),
	}

	for _, node := range params.Nodes {
		request.ActionGroups = append(request.ActionGroups,
			&Ydb_Maintenance.ActionGroup{
				Actions: []*Ydb_Maintenance.Action{
					{
						Action: &Ydb_Maintenance.Action_LockAction{
							LockAction: &Ydb_Maintenance.LockAction{
								Scope: &Ydb_Maintenance.ActionScope{
									Scope: &Ydb_Maintenance.ActionScope_NodeId{
										NodeId: node.NodeId,
									},
								},
								Duration: params.Duration,
							},
						},
					},
				},
			},
		)
	}

	result := &Ydb_Maintenance.MaintenanceTaskResult{}
	_, err := c.ExecuteMaintenanceMethod(result,
		func(ctx context.Context, cl Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error) {
			c.logger.Debug("Invoke CreateMaintenanceTask method")
			return cl.CreateMaintenanceTask(ctx, request)
		},
	)
	if err != nil {
		return result, err
	}
	return result, nil
}

func (c *Cms) RefreshMaintenanceTask(taskID string) (MaintenanceTask, error) {
	result := Ydb_Maintenance.MaintenanceTaskResult{}
	_, err := c.ExecuteMaintenanceMethod(&result,
		func(ctx context.Context, cl Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error) {
			c.logger.Debug("Invoke RefreshMaintenanceTask method")
			return cl.RefreshMaintenanceTask(ctx, &Ydb_Maintenance.RefreshMaintenanceTaskRequest{
				OperationParams: c.f.OperationParams(),
				TaskUid:         taskID,
			})
		},
	)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *Cms) DropMaintenanceTask(taskID string) (string, error) {
	op, err := c.ExecuteMaintenanceMethod(nil,
		func(ctx context.Context, cl Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error) {
			c.logger.Debug("Invoke DropMaintenanceTask method")
			return cl.DropMaintenanceTask(ctx, &Ydb_Maintenance.DropMaintenanceTaskRequest{
				OperationParams: c.f.OperationParams(),
				TaskUid:         taskID,
			})
		},
	)
	if err != nil {
		return "", err
	}

	return op.Status.String(), nil
}

func (c *Cms) CompleteAction(actionIds []*Ydb_Maintenance.ActionUid) (*Ydb_Maintenance.ManageActionResult, error) {
	result := Ydb_Maintenance.ManageActionResult{}
	op, err := c.ExecuteMaintenanceMethod(&result,
		func(ctx context.Context, cl Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error) {
			c.logger.Debug("Invoke CompleteAction method")
			return cl.CompleteAction(ctx, &Ydb_Maintenance.CompleteActionRequest{
				OperationParams: c.f.OperationParams(),
				ActionUids:      actionIds,
			})
		},
	)
	_ = op
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Cms) ExecuteMaintenanceMethod(
	out proto.Message,
	method func(context.Context, Ydb_Maintenance_V1.MaintenanceServiceClient) (OperationResponse, error),
) (*Ydb_Operations.Operation, error) {
	ctx, cancel, err := c.f.ContextWithAuth()
	if err != nil {
		return nil, err
	}
	defer cancel()

	op, err := WrapWithRetries(c.f.GetRetryNumber(), func() (*Ydb_Operations.Operation, error) {
		cc, err := c.f.Connection()
		if err != nil {
			return nil, err
		}

		cl := Ydb_Maintenance_V1.NewMaintenanceServiceClient(cc)
		r, err := method(ctx, cl)
		if err != nil {
			c.logger.Debugf("Invocation error: %+v", err)
			return nil, err
		}
		op := r.GetOperation()
		LogOperation(c.logger, op)
		return op, nil
	})
	if err != nil {
		return nil, err
	}

	if out == nil {
		return op, nil
	}

	if err := op.Result.UnmarshalTo(out); err != nil {
		return op, err
	}

	if op.Status != Ydb.StatusIds_SUCCESS {
		return op, fmt.Errorf("unsuccessful status code: %s", op.Status)
	}

	return op, nil
}

func (c *Cms) ExecuteCMSMethod(
	out proto.Message,
	method func(context.Context, Ydb_Cms_V1.CmsServiceClient) (OperationResponse, error),
) (*Ydb_Operations.Operation, error) {
	ctx, cancel, err := c.f.ContextWithAuth()
	if err != nil {
		return nil, err
	}
	defer cancel()

	op, err := WrapWithRetries(c.f.GetRetryNumber(), func() (*Ydb_Operations.Operation, error) {
		cc, err := c.f.Connection()
		if err != nil {
			return nil, err
		}

		cl := Ydb_Cms_V1.NewCmsServiceClient(cc)
		r, err := method(ctx, cl)
		if err != nil {
			c.logger.Debugf("Invocation error: %+v", err)
			return nil, err
		}
		op := r.GetOperation()
		LogOperation(c.logger, op)
		return op, nil
	})
	if err != nil {
		return nil, err
	}

	if out == nil {
		return op, nil
	}

	if err := op.Result.UnmarshalTo(out); err != nil {
		return op, err
	}

	if op.Status != Ydb.StatusIds_SUCCESS {
		return op, fmt.Errorf("unsuccessful status code: %s", op.Status)
	}

	return op, nil
}