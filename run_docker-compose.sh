#!/bin/bash

echo "Building and Starting Docker Containers..."

# สั่งรัน Docker Compose พร้อม Build ใหม่ทุกครั้ง
docker compose -f docker-compose.yml build
docker compose -f docker-compose.yml --compatibility up -d