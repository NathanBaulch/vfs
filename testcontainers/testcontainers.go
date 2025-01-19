package testcontainers

import "github.com/testcontainers/testcontainers-go"

func withName(name string) testcontainers.ContainerCustomizer {
	return testcontainers.CustomizeRequest(testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{Name: name}})
}
