package extension

import (
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
	"os"
	"path/filepath"
)

type DockerExtension struct {
	cfg *DockerExtensionConf
}

type DockerExtensionConf struct {
	Image string
}

func NewDockerExtension(cfg *DockerExtensionConf) (*DockerExtension, error) {
	err := startDocker(cfg)
	if err != nil {
		return nil, err
	}

	return &DockerExtension{
		cfg: cfg,
	}, err

}

func startDocker(cfg *DockerExtensionConf) error {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		return err
	}
	ctx := context.Background()
	reader, err := cli.ImagePull(ctx, cfg.Image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	io.Copy(os.Stdout, reader)
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: filepath.Base(cfg.Image),
	}, nil, nil, nil, "")
	if err != nil {
		return err
	}
	err = cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}
	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	return nil
}
