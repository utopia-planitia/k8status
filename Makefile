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
		