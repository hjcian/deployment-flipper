GO_CMD=go


rundev:
	air -c cmd/localdev/.air.toml
# $(GO_CMD) run cmd/localdev/main.go
tidy:
	go mod tidy

KC?=kubectl

apply:
	$(KC) apply -f k8s/localdev

delete:
	$(KC) delete -f k8s/localdev

logs:
	$(KC) logs $(shell $(KC) get pods | grep "deploy.*Running" | cut -d" " -f1) -f