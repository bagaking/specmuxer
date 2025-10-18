export CODEX_HOME := $(PWD)/.codex

.PHONY: codex codex_init cc_codex

codex: 
	@echo "Running codex with =$(CODEX_HOME)"
	codex --dangerously-bypass-approvals-and-sandbox

cc_codex: 
	@echo "Running codex with =$(CODEX_HOME)"
	ccmodel --env CODEX_HOME=$(CODEX_HOME) exec run codex -- --dangerously-bypass-approvals-and-sandbox

codex_init:
	specify init . --ai codex