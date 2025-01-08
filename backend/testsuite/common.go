package testsuite

import (
	"os"

	"github.com/c2fo/vfs/v6"
	"github.com/c2fo/vfs/v6/backend/azure"
	"github.com/c2fo/vfs/v6/backend/ftp"
	"github.com/c2fo/vfs/v6/backend/gs"
	"github.com/c2fo/vfs/v6/backend/mem"
	_os "github.com/c2fo/vfs/v6/backend/os"
	"github.com/c2fo/vfs/v6/backend/s3"
	"github.com/c2fo/vfs/v6/backend/sftp"
	"github.com/c2fo/vfs/v6/utils"
)

func CopyOsLocation(loc vfs.Location) vfs.Location {
	ret := utils.Ptr(*loc.(*_os.Location))

	// setup os location
	exists, err := ret.Exists()
	if err != nil {
		panic(err)
	}
	if !exists {
		err := os.Mkdir(ret.Path(), 0o750)
		if err != nil {
			panic(err)
		}
	}

	return ret
}

func CopyMemLocation(loc vfs.Location) vfs.Location {
	return utils.Ptr(*loc.(*mem.Location))
}

func CopyS3Location(loc vfs.Location) vfs.Location {
	return utils.Ptr(*loc.(*s3.Location))
}

func CopySFTPLocation(loc vfs.Location) vfs.Location {
	return utils.Ptr(*loc.(*sftp.Location))
}

func CopyFTPLocation(loc vfs.Location) vfs.Location {
	return utils.Ptr(*loc.(*ftp.Location))
}

func CopyGSLocation(loc vfs.Location) vfs.Location {
	return utils.Ptr(*loc.(*gs.Location))
}

func CopyAzureLocation(loc vfs.Location) vfs.Location {
	return utils.Ptr(*loc.(*azure.Location))
}
