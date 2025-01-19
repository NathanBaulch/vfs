package testcontainers

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"

	"github.com/c2fo/vfs/v6/backend"
	"github.com/c2fo/vfs/v6/backend/sftp"
)

const (
	atmozPort     = "22/tcp"
	atmozUsername = "dummy"
	atmozPassword = "dummy"
)

func registerAtmoz(t *testing.T) string {
	ctx := context.Background()
	is := require.New(t)

	req := testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Name:       "vfs_atmoz",
			Image:      "atmoz/sftp:alpine",
			Env:        map[string]string{"SFTP_USERS": fmt.Sprintf("%s:%s:::upload", atmozUsername, atmozPassword)},
			WaitingFor: wait.ForListeningPort(atmozPort),
		},
		Started: true,
	}
	ctr, err := testcontainers.GenericContainer(ctx, req)
	testcontainers.CleanupContainer(t, ctr)
	is.NoError(err)

	host, err := ctr.Host(ctx)
	is.NoError(err)

	port, err := ctr.MappedPort(ctx, atmozPort)
	is.NoError(err)

	authority := fmt.Sprintf("sftp://%s@%s:%s/upload/", atmozUsername, host, port.Port())
	backend.Register(authority, sftp.NewFileSystem().WithOptions(sftp.Options{Password: vsftpdPassword, KnownHostsCallback: ssh.InsecureIgnoreHostKey()}))
	return authority
}
