// Copyright (C) Alt Research Ltd. All Rights Reserved.
//
// This source code is licensed under the limited license found in the LICENSE file
// in the root directory of this source tree.

package dockerutil

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/alt-research/operator-kit/must"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type ContainerOutput struct {
	Stdout       []byte
	Stderr       []byte
	StatusCode   int64
	ErrorMessage string
}

func (c *ContainerOutput) Error() string {
	return c.ErrorMessage
}

type ContainerOpts struct {
	Name          string
	Networks      []string
	Aliases       []string
	Binds         []string
	Entrypoint    []string
	Cmd           []string
	Env           []string
	WorkDir       string
	Rm            bool
	EndingPhrases []string
	Timeout       time.Duration
}

func SimpleRun(ctx context.Context, cli *client.Client, imageRef string, opts ContainerOpts) (*ContainerOutput, error) {
	log := log.FromContext(ctx)
	var netCfg network.NetworkingConfig
	if os.Getenv("DOCKER_RUN_NETWORK") != "" {
		netCfg = network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				os.Getenv("DOCKER_RUN_NETWORK"): {},
			},
		}
	}
	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{
			Image:      imageRef,
			Env:        opts.Env,
			Cmd:        opts.Cmd,
			Entrypoint: opts.Entrypoint,
			WorkingDir: opts.WorkDir,
			User:       "root",
		},
		&container.HostConfig{Binds: opts.Binds},
		&netCfg,
		&v1.Platform{},
		opts.Name,
	)
	if err != nil {
		return nil, err
	}
	if opts.Rm {
		defer func() {
			if err := cli.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{RemoveVolumes: true, Force: true}); err != nil {
				log.Error(err, "Error removing container", "containerID", resp.ID)
			}
		}()
	}
	rst := &ContainerOutput{}
	outCh, errCh := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNextExit)
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return nil, err
	}

	reader, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true, Follow: true})
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	readErrch := make(chan error, 1)
	defer close(readErrch)
	waitCh := make(chan struct{})
	go func() {
		if _, err := stdcopy.StdCopy(stdout, stderr, reader); err != nil {
			readErrch <- err
		}
		rst.Stdout = stdout.Bytes()
		rst.Stderr = stderr.Bytes()
		defer close(waitCh)
	}()

	dur := must.Default(opts.Timeout, 15*time.Minute)
	timer := time.NewTimer(dur)
	select {
	case err = <-errCh: // container exited with error
	case out := <-outCh: // container exited normally
		if out.Error != nil {
			rst.ErrorMessage = out.Error.Message
		}
		rst.StatusCode = out.StatusCode
	case err = <-readErrch: // error reading container logs
	case <-timer.C: // Total exec timeout
		err = fmt.Errorf("timeout waiting for container %s to exit", resp.ID)
	}
	<-waitCh
	if rst.StatusCode != 0 {
		if rst.ErrorMessage == "" {
			err = fmt.Errorf("container exited with status %d: \n%s", rst.StatusCode, string(rst.Stderr))
		}
	}
	return rst, err
}
