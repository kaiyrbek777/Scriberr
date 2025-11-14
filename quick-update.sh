#!/bin/bash
# Быстрое обновление Go кода в запущенном контейнере БЕЗ полного rebuild

set -e

echo "🔄 Скачиваем обновленные Go файлы..."
curl -s -o internal/api/handlers.go https://raw.githubusercontent.com/kaiyrbek777/Scriberr/claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5/internal/api/handlers.go
curl -s -o internal/transcription/unified_service.go https://raw.githubusercontent.com/kaiyrbek777/Scriberr/claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5/internal/transcription/unified_service.go
curl -s -o internal/transcription/registry/registry.go https://raw.githubusercontent.com/kaiyrbek777/Scriberr/claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5/internal/transcription/registry/registry.go

echo "🔨 Компилируем новый бинарник..."

# Определяем правильный путь для Windows/Linux
if [[ "$OSTYPE" == "msys" || "$OSTYPE" == "win32" ]]; then
  # Windows - используем WSL path формат
  WIN_PATH=$(pwd)
  # Преобразуем /c/Users/... в /c/Users/...
  DOCKER_PATH=$(echo "$WIN_PATH" | sed 's|^/\([a-z]\)/|\L\1:/|')
else
  # Linux/Mac
  DOCKER_PATH=$(pwd)
fi

# Используем тот же Go образ что и в Dockerfile
docker run --rm \
  -v "${DOCKER_PATH}:/src" \
  -w /src \
  golang:1.24-bookworm \
  sh -c "go mod download && CGO_ENABLED=0 go build -o /src/scriberr cmd/server/main.go"

echo "📦 Копируем бинарник в контейнер..."
CONTAINER_ID=$(docker ps -q -f name=scriberr-scriberr)
if [ -z "$CONTAINER_ID" ]; then
  echo "❌ Контейнер не запущен! Запустите: docker-compose up -d"
  exit 1
fi

docker cp ./scriberr $CONTAINER_ID:/app/scriberr

echo "🔄 Перезапускаем контейнер..."
docker-compose restart

echo "✅ Обновление завершено! Проверьте: http://localhost:8080"
rm -f ./scriberr
