package dockerutil

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/docker/docker/client"
	"github.com/kataras/go-fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleRun(t *testing.T) {
	temp, err := os.MkdirTemp("", "test-simple-run-*")
	defer os.RemoveAll(temp)
	require.NoError(t, err)
	cli, err := client.NewClientWithOpts(client.FromEnv)
	require.NoError(t, err)
	out, err := SimpleRun(context.Background(), cli, "busybox", ContainerOpts{
		Binds:      []string{temp + ":/temp"},
		Entrypoint: []string{"sh", "-c"},
		Cmd:        []string{"echo test>/temp/test; ls /temp"},
		Rm:         true,
	})
	require.NoError(t, err)
	assert.Equal(t, []byte("test"), bytes.TrimSpace(out.Stdout))
	assert.EqualValues(t, 0, out.StatusCode)
	assert.True(t, fs.DirectoryExists(temp+"/test"))
}
