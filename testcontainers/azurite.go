package testcontainers

import (
	"context"
	"net/url"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/service"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/azurite"

	"github.com/c2fo/vfs/v6/backend"
	"github.com/c2fo/vfs/v6/backend/azure"
)

func registerAzurite(t *testing.T) string {
	ctx := context.Background()
	is := require.New(t)

	ctr, err := azurite.Run(ctx, "mcr.microsoft.com/azure-storage/azurite:latest", withName("vfs_azurite"))
	testcontainers.CleanupContainer(t, ctr)
	is.NoError(err)

	ep, err := ctr.ServiceURL(ctx, azurite.BlobService)
	is.NoError(err)

	cred, err := service.NewSharedKeyCredential(azurite.AccountName, azurite.AccountKey)
	is.NoError(err)

	u, err := url.JoinPath(ep, azurite.AccountName)
	is.NoError(err)

	cli, err := service.NewClientWithSharedKeyCredential(u, cred, nil)
	is.NoError(err)

	_, err = cli.CreateContainer(ctx, "azurite", nil)
	is.NoError(err)

	backend.Register("https://azurite/", azure.NewFileSystem().WithClient(cli).WithOptions(azure.Options{ServiceURL: u}))
	return "https://azurite/"
}
