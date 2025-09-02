GO ?= go

build:
	$(GO) build

clean:
	$(GO) clean -modcache

test:
	$(GO) clean -testcache
	$(GO) test -v -cover ./...

update-test-snapshots:
	$(GO) clean -testcache
	UPDATE_SNAPSHOT="deployment.yaml/*" $(GO) test -v -cover ./...

lint:
	golangci-lint run
