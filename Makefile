.PHONY: lint
lint:
	docker run \
		-ti \
		--rm \
		-w ${PWD} \
		-v ${PWD}:${PWD} \
		--env GOFLAGS=-buildvcs=false \
		golangci/golangci-lint:v1.48.0-alpine \
		golangci-lint --timeout=540s run ./...

# renovate: datasource=docker depName=docker.io/library/golang
GOLANG_VERSION = 1.24.5-alpine
.PHONY: test
test:
	docker run \
		--volume ${PWD}:/workdir:ro \
		--workdir /workdir \
		docker.io/library/golang:$(GOLANG_VERSION) \
		go test ./...
