package batflow

import (
	"time"

	"go.temporal.io/sdk/workflow"
)

func StartContainerWorkflow(ctx workflow.Context, container *Container) error {
	ao := workflow.ActivityOptions{
		StartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("Container workflow started", "name", container.ID)

	var a *SlurmApptainerActivities
	err := workflow.ExecuteActivity(ctx, a.Start, container).Get(ctx, nil)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return err
	}

	logger.Info("Start container workflow completed.")

	return nil
}
