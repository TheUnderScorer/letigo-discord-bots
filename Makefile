build_base_docker:
	docker buildx build --platform linux/amd64,linux/arm64 -f ./docker/base/Dockerfile -t letigo-discord-bots-base:latest .

build_app:
	docker buildx build --platform linux/amd64,linux/arm64 -f ./docker/server/Dockerfile -t schemik/letigo-discord-bots:latest .