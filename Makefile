GOENV=CGO_ENABLED=0 GOFLAGS="-count=1"
GOCMD=$(GOENV) go
GOTEST=$(GOCMD) test -covermode=atomic -coverprofile=./coverage.out -timeout=20m

rundev:
	@air -c cmd/localdev/.air.toml

tidy:
	@go mod tidy

target=""
test:
	$(GOTEST) -v ./... -run=${target}

see-coverage:
	@go tool cover -html=coverage.out

ci-local-test: test
	@go tool cover -func ./coverage.out

KC?=kubectl

apply:
	$(KC) apply -f k8s/localdev

delete:
	$(KC) delete -f k8s/localdev

logs:
	$(KC) logs $(shell $(KC) get pods | grep "deploy.*Running" | cut -d" " -f1) -f