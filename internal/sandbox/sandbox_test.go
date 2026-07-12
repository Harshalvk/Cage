package sandbox

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockDockerClient struct {
	mock.Mock
}

func (m *MockDockerClient) ImagePull(ctx context.Context, ref string, options image.PullOptions) (io.ReadCloser, error) {
	args := m.Called(ctx, ref, options)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *v1.Platform, containerName string) (container.CreateResponse, error) {
	args := m.Called(ctx, config, hostConfig, networkingConfig, platform, containerName)
	return args.Get(0).(container.CreateResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error {
	args := m.Called(ctx, containerID, options)
	return args.Error(0)
}

func (m *MockDockerClient) ContainerExecCreate(ctx context.Context, containerID string, config container.ExecOptions) (container.ExecCreateResponse, error) {
	args := m.Called(ctx, containerID, config)
	return args.Get(0).(container.ExecCreateResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerExecAttach(ctx context.Context, execID string, config container.ExecStartOptions) (types.HijackedResponse, error) {
	args := m.Called(ctx, execID, config)
	return args.Get(0).(types.HijackedResponse), args.Error(1)
}

func (m *MockDockerClient) ContainerExecInspect(ctx context.Context, execID string) (container.ExecInspect, error) {
	args := m.Called(ctx, execID)
	return args.Get(0).(container.ExecInspect), args.Error(1)
}

func (m *MockDockerClient) ContainerInspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	args := m.Called(ctx, containerID)
	return args.Get(0).(container.InspectResponse), args.Error(1)
}

func (m *MockDockerClient) CopyToContainer(ctx context.Context, containerID, path string, content io.Reader, options container.CopyToContainerOptions) error {
	args := m.Called(ctx, containerID, path, content, options)
	return args.Error(0)
}

func (m *MockDockerClient) CopyFromContainer(ctx context.Context, containerID, srcPath string) (io.ReadCloser, container.PathStat, error) {
	args := m.Called(ctx, containerID, srcPath)
	return args.Get(0).(io.ReadCloser), args.Get(1).(container.PathStat), args.Error(2)
}

func TestCreateSandbox_Success(t *testing.T) {
	mockClient := new(MockDockerClient)

	mockClient.On("ImagePull", mock.Anything, "alpine:latest", mock.Anything).Return(io.NopCloser(strings.NewReader("")), nil)

	mockClient.On("ContainerCreate", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(container.CreateResponse{ID: "fake-container-id"}, nil)

	mockClient.On("ContainerStart", mock.Anything, "fake-container-id", mock.Anything).Return(nil)

	sm := NewSandboxManagerWithClient(mockClient)

	id, err := sm.CreateSandbox(context.Background(), "alpine:latest")

	require.NoError(t, err)
	assert.Equal(t, "fake-container-id", id)
	mockClient.AssertExpectations(t)
}
