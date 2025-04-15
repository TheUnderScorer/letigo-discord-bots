#!/bin/bash
set -e

echo "[INFO] Pulling latest image..."
docker-compose pull prod

echo "[INFO] Recreating container..."
docker-compose up -d prod

echo "[INFO] Removing unused images..."
docker image prune -f

echo "[INFO] Bot redeployed!"