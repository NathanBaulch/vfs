module github.com/c2fo/vfs/backend/sftp

go 1.23.2

replace github.com/c2fo/vfs/v6 => ../..

require (
	github.com/c2fo/vfs/v6 v6.20.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/pkg/sftp v1.13.7
	github.com/stretchr/testify v1.9.0
	golang.org/x/crypto v0.28.0
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.2 // indirect
	golang.org/x/sys v0.26.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
