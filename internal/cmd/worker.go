package cmd

import (
	"fmt"
	"os"

	"github.com/batx-dev/batflow"
	"github.com/urfave/cli/v2"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"golang.org/x/crypto/ssh"
)

func getWorkerCommand() *cli.Command {
	cmd := &cli.Command{
		Name:  "worker",
		Usage: "Start batflow worker",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "ssh-address",
				Value:   "127.0.0.1:22",
				Usage:   "The address to ssh",
				EnvVars: []string{"BATFLOW_SSH_ADDRESS"},
			},
			&cli.StringFlag{
				Name:     "ssh-username",
				Value:    "",
				Usage:    "The username to ssh",
				EnvVars:  []string{"BATFLOW_SSH_USERNAME"},
				Required: true,
			},
			&cli.StringFlag{
				Name:    "ssh-keyfile",
				Value:   "",
				Usage:   "The keyfile to ssh",
				EnvVars: []string{"BATFLOW_SSH_KEYFILE"},
			},
		},
		Action: runWorker,
	}
	return cmd

}

func runWorker(ctx *cli.Context) error {
	// Dial to ssh server.
	sshConfig := &ssh.ClientConfig{
		User:            ctx.String("ssh-username"),
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	if keyfile := ctx.String("ssh-keyfile"); keyfile != "" {
		key, err := os.ReadFile(keyfile)
		if err != nil {
			return fmt.Errorf("read ssh key file: %v", err)
		}
		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return fmt.Errorf("parse key file: %v", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(signer))
	}
	sshClient, err := ssh.Dial("tcp", ctx.String("ssh-address"), sshConfig)
	if err != nil {
		return fmt.Errorf("ssh dial: %v", err)
	}
	defer sshClient.Close()

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.Dial(client.Options{})
	if err != nil {
		return fmt.Errorf("unable to create client: %v", err)
	}
	defer c.Close()

	w := worker.New(c, "batflow", worker.Options{})

	w.RegisterWorkflow(batflow.StartContainerWorkflow)
	w.RegisterActivity(batflow.NewSlurmApptainerActivites(sshClient))

	err = w.Run(worker.InterruptCh())
	if err != nil {
		return fmt.Errorf("unable to start worker: %v", err)
	}
	return nil
}
