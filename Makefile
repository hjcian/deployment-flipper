GO_CMD=go


rundev:
	air -c cmd/localdev/.air.toml
# $(GO_CMD) run cmd/localdev/main.go
tidy:
	go mod tidy
