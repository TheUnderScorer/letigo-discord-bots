build_base_docker:
	docker buildx build --platform linux/amd64,linux/arm64 -f ./docker/base/Dockerfile -t letigo-discord-bots-base:latest .

build_app_docker:
	docker buildx build --platform linux/amd64,linux/arm64 -f ./docker/server/Dockerfile -t schemik/letigo-discord-bots:latest .

push_app_image:
	docker push schemik/letigo-discord-bots:latest

publish_app_image: build_base_docker build_app_docker push_app_image