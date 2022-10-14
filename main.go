package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type RunOption struct {
	TargetVersion string
	Command       string
}

func main() {

	args := os.Args[1:]

	log.Println(args)

	if len(os.Args) < 2 {
		log.Fatalln("required at least two arguments")
	}

	command := os.Args[1]
	targetVersion := os.Args[2]

	log.Printf("command = %s\n", command)
	log.Printf("target version = %s\n", targetVersion)

	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	newImage := fmt.Sprintf("docker.io/oscarzhou/ubuntu:%s", targetVersion)
	reader, err := cli.ImagePull(ctx, newImage, types.ImagePullOptions{})
	if err != nil {
		panic(err)
	}
	io.Copy(os.Stdout, reader)

	// stop the old agent container

	// create the new container
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: "oscarzhou/ubuntu:2",
		// Cmd:   []string{"cat", "version.txt", "&&", "cd", "/var/run", "&&", "ls"},
		Cmd: []string{"ls", "/var/run"},
	}, &container.HostConfig{
		Binds: []string{
			"/var/run/docker.sock:/var/run/docker.sock",
		},
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	statusCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		panic(err)
	}

	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

}
