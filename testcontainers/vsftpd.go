package testcontainers

import (
	"context"
	"fmt"
	"testing"

	"github.com/c2fo/vfs/v6/utils"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/c2fo/vfs/v6/backend"
	"github.com/c2fo/vfs/v6/backend/ftp"
)

const (
	vsftpdPort     = "21/tcp"
	vsftpdPassword = "dummy"
)

func registerVSFTPC(t *testing.T) string {
	ctx := context.Background()
	is := require.New(t)

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:         "vfs_vsftpd",
			Image:        "fauria/vsftpd:latest",
			ExposedPorts: []string{"21", "21100-21110:21100-21110"},
			Env:          map[string]string{"FTP_PASS": vsftpdPassword},
			WaitingFor:   wait.ForListeningPort(vsftpdPort),
		},
		Started: true,
	}
	ctr, err := testcontainers.GenericContainer(ctx, req)
	testcontainers.CleanupContainer(t, ctr)
	is.NoError(err)

	host, err := ctr.Host(ctx)
	is.NoError(err)

	port, err := ctr.MappedPort(ctx, vsftpdPort)
	is.NoError(err)

	authority := fmt.Sprintf("ftp://admin@%s:%s/", host, port.Port())
	backend.Register(authority, ftp.NewFileSystem().WithOptions(ftp.Options{Password: vsftpdPassword, WritingMDTM: utils.Ptr(true)}))
	return authority
}
