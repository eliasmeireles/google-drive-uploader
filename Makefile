.PHONY: buildx buildx-build push run-test build test-uploader build-docker-test test-docker debug-docker

# Multi-platform Docker build commands
# Setup Buildx builder
buildx:
	@docker buildx create --name buildxBuilder --use || true
	@docker buildx inspect buildxBuilder --bootstrap

# Build multi-platform image and push
buildx-build:
	@read -p "Enter the tag version: " TAG; \
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/eliasmeireles/cli/google-drive-uploader:$$TAG --push .

run-test:
	@go test -v ./...

build:
	@go build -v ./cmd/uploader
	@./uploader --help

test-uploader:
	@echo "Testing uploader with smart organization"; \
	read -p "Enter the root folder ID: " ROOT_FOLDER_ID; \
	go run ./cmd/uploader --smart-organize --root-folder-id "$$ROOT_FOLDER_ID" --folder-name GDU_CLI_TEST ./test/myFakeDatabase_backup_20251224_084205.txt

build-docker-test:
	@echo "Building docker image for testing..."
	@docker build -f Dockerfile.test -t google-drive-uploader:test .

test-docker:
	@echo "Testing docker image with smart organization"; \
	read -p "Enter the root folder ID: " ROOT_FOLDER_ID; \
	docker run --rm \
		-v $(PWD)/.out:/etc/google-drive-uploader \
		-v $(PWD)/test:/data \
		ghcr.io/eliasmeireles/cli/google-drive-uploader:latest \
		--smart-organize \
		--root-folder-id "$$ROOT_FOLDER_ID" \
		--folder-name GDU_CLI_TEST_DOCKER \
		--token-path /etc/google-drive-uploader/token.json \
		/data/myFakeDatabase_backup_20251224_084205.txt

debug-docker: build-docker-test
	@docker run --rm -it \
		-v $(PWD)/.out:/etc/google-drive-uploader \
		-v $(PWD)/test:/data \
		--entrypoint sh \
		google-drive-uploader:test
