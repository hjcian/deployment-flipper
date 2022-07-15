GO_CMD=go


rundev:
	$(GO_CMD) run cmd/localdev/main.go

tidy:
	go mod tidy