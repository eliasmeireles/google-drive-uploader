.PHONY: buildx buildx-build push

# Multi-platform Docker build commands
# Setup Buildx builder
buildx:
	@docker buildx create --name buildxBuilder --use || true
	@docker buildx inspect buildxBuilder --bootstrap

# Build multi-platform image and push
buildx-build:
	@read -p "Enter the tag version: " TAG; \
	docker buildx build --platform linux/amd64,linux/arm64 -t ghcr.io/eliasmeireles/cli/google-driver-uploader:$$TAG --push .
