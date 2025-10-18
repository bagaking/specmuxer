export CODEX_HOME := $(PWD)/.codex

.PHONY: codex codex_init cc_codex lint test coverage

codex: 
	@echo "Running codex with =$(CODEX_HOME)"
	codex --dangerously-bypass-approvals-and-sandbox

cc_codex: 
	@echo "Running codex with =$(CODEX_HOME)"
	ccmodel --env CODEX_HOME=$(CODEX_HOME) exec run codex -- --dangerously-bypass-approvals-and-sandbox

codex_init:
	specify init . --ai codex

lint:
	@command -v golangci-lint >/dev/null || (echo "golangci-lint is required for lint target" >&2 && exit 1)
	golangci-lint run ./...

test:
	go test ./... -race -count=1

coverage:
	go test ./... -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
	@go tool cover -func=coverage.out | tail -n1 | awk '{print $$3}' | awk -F% '{if ($$1 < 95) {printf \"coverage %.2f%% is below required 95%%\\n\", $$1; exit 1}}'
