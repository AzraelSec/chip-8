.PHONY: help
## help: prints this message
help:
	@echo "usage:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

.PHONY: tidy
## tidy: format files and tidy module dependencies
tidy:
	@echo "Formatting .go files"
	go fmt ./...
	@echo "Tidying module dependencies"
	go mod tidy

.PHONY: build
## build: build the interpreter binary
build:
	go build .
