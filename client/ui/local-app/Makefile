.PHONY: *
.DEFAULT_GOAL:=help

help:
	@echo "Commands available"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /' | sort

## check: Verify code is properly formatted
check:
	npm run format
	npm run lint
	npm run check

## build: Build the UI
build:
	npm run build

## dev: Run the UI in development mode
dev:
	npm run dev

## deps: Install dependencies (with pnpm)
deps:
	pnpm i

## updates: List available updates for direct dependencies
updates:
	pnpm outdated
