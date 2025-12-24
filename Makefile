.PHONY: buildx buildx-build push run-test build test-uploader

# Multi-platform Docker build commands
# Setup Buildx builder
buildx:
	@docker buildx create --name buildxBuilder --use || true
	@docker buildx inspect buildxBuilder --bootstrap

# Build multi-platform image and push
buildx-build:
	@read -p "Enter the tag version: " TAG; \
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/eliasmeireles/cli/google-driver-uploader:$$TAG --push .

run-test:
	@go test -v ./...

build:
	@go build -v ./cmd/uploader
	@./uploader --help

test-uploader:
	@echo "Testing uploader with smart organization"; \
	read -p "Enter the root folder ID: " ROOT_FOLDER_ID; \
	go run ./cmd/uploader --smart-organize --root-folder-id "$$ROOT_FOLDER_ID" --folder-name GDU_CLI_TEST ./test/myFakeDatabase_backup_20251224_084205.txt
