.PHONY: build_base_dockerx build_base_docker build_wojciech_bot_docker build_wojciech_bot_dockerx push_wojciech_bot_app_image publish_app_image deploy_wojciech_bot

build_base_dockerx:
	docker buildx build --load --platform linux/amd64,linux/arm64 -f ./docker/base/Dockerfile -t letigo-discord-bots-base:latest .

build_base_docker:
	docker build -f ./docker/base/Dockerfile -t letigo-discord-bots-base:latest .

build_wojciech_bot_dockerx:
	docker buildx build --platform linux/amd64,linux/arm64 -f ./docker/wojciech-bot/Dockerfile -t schemik/letigo-discord-bots-wojciech:latest .

build_wojciech_bot_docker:
	docker build -f ./docker/wojciech-bot/Dockerfile -t schemik/letigo-discord-bots-wojciech:latest .

push_wojciech_bot_app_image:
	docker push schemik/letigo-discord-bots:latest

publish_wojciech_bot_image: build_base_dockerx build_wojciech_bot_dockerx push_wojciech_bot_app_image

deploy_wojciech_bot: 
	@echo "[INFO] Pulling latest image..."
	docker compose pull wojciech-bot-prod
	@echo "[INFO] Recreating container..."
	docker compose up -d wojciech-bot-prod
	@echo "[INFO] Removing unused images..."
	docker image prune -f
	@echo "[INFO] Wojciech Bot redeployed!"