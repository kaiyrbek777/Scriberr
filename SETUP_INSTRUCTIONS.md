# 📋 Инструкция по установке и запуску Protocol Transcription Service

Этот документ содержит полную инструкцию по клонированию, настройке и запуску проекта локально.

---

## 🚀 Быстрый старт

### 1. Клонирование репозитория

```bash
# Клонировать репозиторий
git clone https://github.com/kaiyrbek777/Scriberr.git
cd Scriberr

# Переключиться на ветку с изменениями
git checkout claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5
```

### 2. Вариант А: Запуск через Docker (РЕКОМЕНДУЕТСЯ)

**Преимущества:** Не нужно устанавливать Go, Node.js, Python - всё внутри контейнера

```bash
# Запуск с CPU
docker-compose up --build

# ИЛИ запуск с GPU (если есть NVIDIA GPU)
docker-compose -f docker-compose.cuda.yml up --build
```

Сервис будет доступен по адресу: **http://localhost:8080**

### 3. Вариант Б: Запуск локально (для разработки)

#### Требования:
- **Go** 1.21+
- **Node.js** 18+
- **Python** 3.10+ (для транскрибации)

#### Шаги:

```bash
# 1. Установить backend зависимости
go mod download

# 2. Установить frontend зависимости
cd web/frontend
npm install
cd ../..

# 3. Собрать frontend
cd web/frontend
npm run build
cd ../..

# 4. Запустить сервер
go run cmd/server/main.go
```

Сервис будет доступен по адресу: **http://localhost:8080**

---

## 👤 Первый запуск - Регистрация

1. Откройте браузер: **http://localhost:8080**
2. Вы увидите страницу регистрации
3. Создайте первый аккаунт - **он автоматически станет администратором**
4. Последующие пользователи (если разрешите регистрацию) будут обычными пользователями

---

## 🔐 Роли и права доступа

### Администратор (Admin)
**Первый зарегистрированный пользователь**

✅ **Может:**
- Загружать аудио и транскрибировать
- Редактировать транскрипции
- **Создавать и редактировать профили транскрибации** (Settings → Transcription)
- **Создавать и редактировать шаблоны протоколов** (Settings → Summary)
- **Настраивать LLM** (Settings → LLMs)
- **Просматривать статистику** (`GET /api/v1/admin/statistics`)
- Управлять API ключами

### Обычный пользователь (User)
**Все последующие зарегистрированные пользователи**

✅ **Может:**
- Загружать аудио и транскрибировать
- Редактировать транскрипции
- Использовать шаблоны протоколов (созданные админом)
- Изменить пароль и username (Settings → Account)
- Управлять своими API ключами

❌ **НЕ может:**
- Создавать/редактировать профили транскрибации
- Создавать/редактировать шаблоны
- Настраивать LLM
- Просматривать статистику

---

## 🔧 API Endpoints

### Простая транскрибация (cURL-friendly)

```bash
# Получить токен (login)
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "your_password"}'

# Ответ: {"token": "eyJhbGc...", "user": {"id": 1, "username": "admin", "role": "admin"}}

# Транскрибировать аудио
curl -X POST http://localhost:8080/api/v1/transcription/simple \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -F "file=@audio.mp3" \
  -F "language=en" \
  -F "diarization=true"

# Ответ:
# {
#   "success": true,
#   "job_id": "uuid-here",
#   "status": "pending",
#   "message": "Transcription started",
#   "status_url": "/api/v1/transcription/{job_id}/status"
# }

# Проверить статус
curl -X GET http://localhost:8080/api/v1/transcription/{job_id}/status \
  -H "Authorization: Bearer YOUR_TOKEN"

# Получить результат
curl -X GET http://localhost:8080/api/v1/transcription/{job_id}/transcript \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Статистика для администратора

```bash
curl -X GET http://localhost:8080/api/v1/admin/statistics \
  -H "Authorization: Bearer YOUR_ADMIN_TOKEN"

# Ответ:
# {
#   "transcriptions": {
#     "total": 150,
#     "completed": 145,
#     "failed": 5,
#     "recent_7d": 23
#   },
#   "users": { "total": 5 },
#   "templates": { "total": 8 },
#   "summaries": { "total": 120 },
#   "api_keys": { "total": 3, "active": 2 }
# }
```

---

## 🔄 Работа с Git

### Получение последних изменений

```bash
# Убедитесь что вы на правильной ветке
git checkout claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5

# Получить последние изменения
git pull origin claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5
```

### Создание Pull Request

```bash
# 1. Убедитесь что все изменения закоммичены
git status

# 2. Push ветки (если еще не сделано)
git push -u origin claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5

# 3. Создайте PR на GitHub:
# https://github.com/kaiyrbek777/Scriberr/compare/claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5
```

---

## 🗄️ База данных

Проект использует **SQLite** (файл `data/scriberr.db`).

### ⚠️ Важно при первом запуске с новой схемой

Так как добавлено новое поле `role` в модель User, вам нужно:

**Вариант 1: Удалить старую БД (самый простой)**
```bash
rm -rf data/
```

**Вариант 2: Миграция вручную (если нужно сохранить данные)**
```bash
# Установить sqlite3
sudo apt-get install sqlite3  # Ubuntu/Debian
brew install sqlite3          # macOS

# Открыть БД
sqlite3 data/scriberr.db

# Добавить поле role
ALTER TABLE users ADD COLUMN role VARCHAR(20) DEFAULT 'user';

# Сделать первого пользователя админом
UPDATE users SET role = 'admin' WHERE id = 1;

# Выйти
.exit
```

---

## 🐛 Решение проблем

### Проблема: "WhisperX не установлен"

**Решение:** WhisperX теперь загружается автоматически при первой транскрибации (lazy loading). Просто подождите пока Python окружение установится.

### Проблема: "403 Forbidden" при доступе к admin endpoints

**Решение:** Убедитесь что вы залогинены как администратор (первый зарегистрированный пользователь). Проверьте роль в JWT токене:

```bash
# Декодировать JWT (вставьте свой токен)
echo "YOUR_TOKEN_PAYLOAD_PART" | base64 --decode | jq
# Должно быть: "role": "admin"
```

### Проблема: Фронтенд не отображается

**Решение 1:** Пересобрать frontend
```bash
cd web/frontend
npm run build
cd ../..
go run cmd/server/main.go
```

**Решение 2:** Использовать Docker
```bash
docker-compose down
docker-compose up --build
```

### Проблема: База данных заблокирована

**Решение:**
```bash
# Остановить все процессы
docker-compose down
pkill -f "go run cmd/server"

# Удалить lock файл
rm -f data/scriberr.db-shm data/scriberr.db-wal

# Перезапустить
docker-compose up
```

---

## 📊 Swagger API Documentation

После запуска сервера, API документация доступна по адресу:

**http://localhost:8080/swagger/index.html**

---

## 🎯 Основные изменения в этой ветке

1. ✅ **Система ролей**: Admin/User с разграничением прав
2. ✅ **Удаление брендирования Scriberr**
3. ✅ **Lazy loading WhisperX** - быстрый старт без долгой загрузки
4. ✅ **Simple cURL API** - `/api/v1/transcription/simple`
5. ✅ **Admin статистика** - `/api/v1/admin/statistics`
6. ✅ **Role-based UI** - Settings показывает только доступные разделы

---

## 📝 Структура проекта

```
Scriberr/
├── cmd/server/          # Точка входа приложения
├── internal/
│   ├── api/            # HTTP handlers и routing
│   ├── auth/           # JWT authentication
│   ├── database/       # SQLite + GORM
│   ├── models/         # Data models (User, TranscriptionJob, etc)
│   ├── queue/          # Background job processing
│   └── transcription/  # STT processing (WhisperX, etc)
├── pkg/
│   └── middleware/     # Auth middleware (JWT, Admin-only)
├── web/
│   └── frontend/       # React + TypeScript frontend
└── data/               # Database, uploads, Python env
```

---

## 🔗 Полезные ссылки

- **GitHub Repo**: https://github.com/kaiyrbek777/Scriberr
- **Branch**: `claude/remove-scriberr-branding-011CV5kgbiLtaycotVfYTKU5`
- **Issues**: https://github.com/kaiyrbek777/Scriberr/issues

---

## 💡 Советы

1. **Используйте Docker** для быстрого старта
2. **API Keys** удобны для автоматизации (не истекают как JWT)
3. **Дефолтный профиль** (Transcription Settings) применяется ко всем новым транскрибациям
4. **Шаблоны протоколов** (Summary Templates) можно переиспользовать для разных транскрипций

---

**Вопросы?** Создайте issue на GitHub!
