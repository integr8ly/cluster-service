BUILD_TARGET=./

.PHONY: build/cli
build/cli: build/cli/local

.PHONY: build/cli/local
build/cli/local:
	go build -o=$(BUILD_TARGET) ./cmd/cli

.PHONY: code/fix
code/fix:
	@gofmt -w `find . -type f -name '*.go' -not -path "./vendor/*"`

.PHONY: code/check
code/check:
	go vet ./...

.PHONY: code/gen
code/gen:
	go generate ./...

.PHONY: test/unit
test/unit:
	go test -v ./...

.PHONY: vendor/check
vendor/check: vendor/fix
	git diff --exit-code vendor/

.PHONY: vendor/fix
vendor/fix:
	go mod tidy
	go mod vendor

.PHONY: setup/goreleaser
setup/goreleaser:
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh | sh