.PHONY: *
.DEFAULT_GOAL:=help

help:
	@echo "Commands available"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /' | sort

## deployment-images: Build necessary images to deploy an Algorand node
deployment-images:
	docker build --platform linux/amd64 -t ghcr.io/bryk-io/algod:stable -f algod/Dockerfile algod/.
	docker build --platform linux/amd64 -t ghcr.io/bryk-io/algod:beta -f algod/Dockerfile --build-arg channel=beta algod/.
	docker build --platform linux/amd64 -t ghcr.io/bryk-io/algod-indexer:2.15.1 -f algod-indexer/Dockerfile algod-indexer/.
	docker build --platform linux/amd64 -t ghcr.io/bryk-io/ipfs:0.12.2 -f ipfs/Dockerfile ipfs/.
	docker build --platform linux/amd64 -t ghcr.io/bryk-io/algoid-resolver:0.4.0 -f algoid-resolver/Dockerfile algoid-resolver/.
