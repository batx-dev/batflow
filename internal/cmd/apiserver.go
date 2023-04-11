package cmd

import (
	"context"
	"fmt"
	"net/http"

	"github.com/batx-dev/batflow"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	restful "github.com/emicklei/go-restful/v3"
	"github.com/urfave/cli/v2"
	workflowpb "go.temporal.io/api/workflow/v1"
	"go.temporal.io/api/workflowservice/v1"
	"go.temporal.io/sdk/client"
	"k8s.io/apimachinery/pkg/util/rand"
)

func getAPIServerCommand() *cli.Command {
	cmd := &cli.Command{
		Name:  "apiserver",
		Usage: "Run batflow apiserver",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:  "http-listen",
				Value: ":8080",
				Usage: "The address to http listen",
			},
		},
		Action: runAPIServer,
	}
	return cmd

}

func runAPIServer(ctx *cli.Context) error {
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		return fmt.Errorf("dial temporal: %v", err)
	}

	c := restful.NewContainer()
	c.Add(NewContainerResource(temporalClient).WebService())

	if err := http.ListenAndServe(ctx.String("http-listen"), c); err != nil {
		return fmt.Errorf("http listen and serve: %v", err)
	}

	return nil
}

type ContainerResource struct {
	temporalClient client.Client
}

func NewContainerResource(client client.Client) *ContainerResource {
	return &ContainerResource{temporalClient: client}
}

func (r *ContainerResource) WebService() *restful.WebService {
	ws := new(restful.WebService)
	ws.
		Path("/api/v1beta1").
		Consumes(restful.MIME_JSON).
		Produces(restful.MIME_JSON)

	tags := []string{"containers"}

	ws.Route(ws.GET("/containers").To(r.listContainers).
		Doc("list all containers").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Writes([]batflow.Container{}).
		Returns(200, "OK", []batflow.Container{}))

	ws.Route(ws.POST("/containers:start").To(r.startContainer).
		Doc("start a container").
		Metadata(restfulspec.KeyOpenAPITags, tags).
		Reads(batflow.Container{}).
		Writes(batflow.Container{}).
		Returns(201, "Created", batflow.Container{}))

	return ws
}

type ServiceError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func WriteServiceError(res *restful.Response, httpStatus int, err error) {
	var message string
	if err != nil {
		message = err.Error()
	}
	res.WriteHeaderAndEntity(httpStatus, &ServiceError{
		Code:    httpStatus,
		Message: message,
	})
}

func (r *ContainerResource) listContainers(req *restful.Request, res *restful.Response) {
	wfs, err := r.listWorkflows(req.Request.Context())
	if err != nil {
		WriteServiceError(res, http.StatusInternalServerError, fmt.Errorf("list containers: %v", err))
		return
	}

	containers := make([]*batflow.Container, len(wfs))
	for i, wf := range wfs {
		c := presentContainer(wf)
		containers[i] = c
	}

	res.WriteEntity(containers)
}

func (r *ContainerResource) startContainer(req *restful.Request, res *restful.Response) {
	container := new(batflow.Container)
	if err := req.ReadEntity(container); err != nil {
		WriteServiceError(res, http.StatusInternalServerError, fmt.Errorf("read request body: %v", err))
		return
	}
	if container.ID == "" {
		container.ID = rand.String(16)
	}

	_, err := r.temporalClient.ExecuteWorkflow(req.Request.Context(), client.StartWorkflowOptions{
		ID:        container.ID,
		TaskQueue: "batflow",
	},
		batflow.StartContainerWorkflow,
		container,
	)
	if err != nil {
		WriteServiceError(res, http.StatusInternalServerError, fmt.Errorf("start container: %v", err))
		return
	}
	res.WriteHeaderAndEntity(http.StatusCreated, container)
}

func (r *ContainerResource) listWorkflows(ctx context.Context) ([]*workflowpb.WorkflowExecutionInfo, error) {
	var executions []*workflowpb.WorkflowExecutionInfo
	var nextPageToken []byte

	for {
		resp, err := r.temporalClient.ListWorkflow(ctx, &workflowservice.ListWorkflowExecutionsRequest{
			PageSize:      10,
			NextPageToken: nextPageToken,
		})
		if err != nil {
			return executions, err
		}

		executions = append(executions, resp.Executions...)
		if len(resp.NextPageToken) == 0 {
			return executions, nil
		}
		nextPageToken = resp.NextPageToken
	}
}

func presentContainer(we *workflowpb.WorkflowExecutionInfo) *batflow.Container {
	c := new(batflow.Container)

	c.ID = we.Execution.WorkflowId
	c.Status = we.Status.String()

	return c
}
