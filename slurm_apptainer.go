package batflow

import (
	"context"
	"fmt"
	"path"
	"strings"

	"golang.org/x/crypto/ssh"
	"k8s.io/apimachinery/pkg/util/rand"
)

const (
	DeviceNvidiaGPUKey = "nvidia.com/gpu"
)

type Container struct {
	// Name of the container specified as a DNS_LABEL. Each Container must
	// have a unique name.
	Name string `json:"name"`
	// Image container image name.
	Image Image `json:"image"`
	// Command to execute.
	Command []string `json:"command"`
	// Ports list of port to expose from the container.
	Ports []uint16 `json:"ports"`
	// Resources compute resources required by container.
	Resource Resource `json:"resource"`
}

// Image container image name.
type Image struct {
	URI string `json:"uri"`
}

// Resource compute resources required by container.
type Resource struct {
	CPU struct {
		Cores uint `json:"cores"`
	}
	Memory struct {
		Size string `json:"size"`
	}
	Devices map[string]uint `json:"devices"`
}

type SlurmApptainerActivities struct {
	sshClient *ssh.Client
}

func NewSlurmApptainerActivites(client *ssh.Client) *SlurmApptainerActivities {
	return &SlurmApptainerActivities{sshClient: client}
}

// Start a named instance of the given container image
func (a *SlurmApptainerActivities) Start(ctx context.Context, container *Container) (err error) {
	sess, err := a.sshClient.NewSession()
	if err != nil {
		return fmt.Errorf("ssh new session: %v", err)
	}
	defer sess.Close()

	if err := sess.Run(buildSlurmJobSubmitCommand(container)); err != nil {
		return fmt.Errorf("submit job: %v", err)
	}

	return nil
}

// List all running and named container instances
func (a *SlurmApptainerActivities) List(ctx context.Context, names []string) (err error) {
	return
}

func buildSlurmJobSubmitCommand(container *Container) string {
	if container.Name == "" {
		container.Name = rand.String(16)
	}
	workDir := path.Join("~/.batflow/jobs", container.Name)

	return `sh <<- 'CMD'
		mkdir -p ` + workDir + ` || exit $?
		cd ` + workDir + ` || exit $?

		# Generate job script.
		cat <<- 'SCRIPT' > job.sh || exit $?
			#!/usr/bin/env bash
			` + buildSlurmJobOptions(container) + `

			[[ -f ~/.batflow/env ]] && source ~/.batflow/env || true
			` + buildSlurmJobStepCommand(container) + `
		SCRIPT

		# Submit job script.
		sbatch job.sh
	CMD`
}

func buildSlurmJobOptions(container *Container) string {
	opts := []string{
		fmt.Sprintf("--job-name=%q", container.Name),
	}
	if count, ok := container.Resource.Devices[DeviceNvidiaGPUKey]; ok {
		opts = append(opts, fmt.Sprintf("--gpus=%d", count))
	}
	for i, opt := range opts {
		opts[i] = "#SBATCH " + opt
	}
	return strings.Join(opts, "\n")
}

func buildSlurmJobStepCommand(container *Container) string {
	command := []string{
		"apptainer",
		"exec",
		"--compat",
		"--fakeroot",
	}
	if _, ok := container.Resource.Devices[DeviceNvidiaGPUKey]; ok {
		command = append(command, "--nv")
	}
	command = append(command, container.Image.URI)
	command = append(command, container.Command...)
	return strings.Join(command, " ")
}
