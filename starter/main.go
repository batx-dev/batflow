package main

import (
	"context"
	"log"

	"github.com/batx-dev/batflow"
	"go.temporal.io/sdk/client"
)

func main() {
	// The client is a heavyweight object that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	workflowOptions := client.StartWorkflowOptions{
		TaskQueue: "batflow",
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, batflow.StartContainerWorkflow,
		&batflow.Container{
			Image: batflow.Image{
				URI: "docker://jupyter/minimal-notebook",
			},
			Command: []string{"jupyter", "notebook", "--port", "8888", "--no-browser"},
			Ports:   []uint16{8888},
			Resource: batflow.Resource{
				Devices: map[string]uint{
					batflow.DeviceNvidiaGPUKey: 1,
				},
			},
		},
	)
	if err != nil {
		log.Fatalln("Unable to execute workflow", err)
	}

	log.Println("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	err = we.Get(context.Background(), nil)
	if err != nil {
		log.Fatalln("Unable get workflow result", err)
	}
}
