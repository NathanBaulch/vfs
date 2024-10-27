MOCKERIES := $(shell find . -type f -name '.mockery.yaml' -exec dirname {} \; | sort)
SUBMODULES := $(shell find . -type f -name 'go.mod' -exec dirname {} \; | sort)

mockery:
	set -e; for dir in $(MOCKERIES); do \
	  echo "go test in $${dir}"; \
	  (cd "$${dir}" && mockery); \
	done

test:
	set -e; for dir in $(SUBMODULES); do \
	  echo "go test in $${dir}"; \
	  (cd "$${dir}" && go test); \
	done

build:
	set -e; for dir in $(SUBMODULES); do \
	  echo "go build in $${dir}"; \
	  (cd "$${dir}" && go build); \
	done

upgrade:
	set -e; for dir in $(SUBMODULES); do \
	  echo "go upgrade in $${dir}"; \
	  (cd "$${dir}" && go get -u -t ./...); \
	done
