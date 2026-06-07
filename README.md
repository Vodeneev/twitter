# Yapper

**Yapper** — клон Twitter (микроблог) с полным базовым функционалом. Название
в стиле Yandex/Yango (`Ya-` + «yap», болтать); пост называется **yap**.

Архитектура повторяет «базовую схему» референс-проекта: Go + chi + pgx +
goose на бэкенде, Next.js 14 + TypeScript + Tailwind + next-intl на фронте,
сессии в БД + HttpOnly-cookie, письма через SMTP (Resend в проде), медиа в
S3-совместимом хранилище (MinIO локально, Cloudflare R2 в проде), Caddy как
reverse-proxy на один origin, деплой на Yandex Cloud VM, Cloudflare сверху как
DNS/CDN.

## Стек

| Слой           | Выбор                                                       |
| -------------- | ----------------------------------------------------------- |
| Backend        | Go 1.25 + `chi` + `pgx/v5` + `goose` + `coder/websocket`    |
| База данных    | PostgreSQL 16                                               |
| Frontend       | Next.js 14 (App Router) + TypeScript + Tailwind             |
| i18n           | `next-intl` — русский (по умолчанию) + английский           |
| Почта          | SMTP-транспорт; в проде Resend (`smtp.resend.com`)          |
| Медиа          | S3-совместимое: MinIO (dev) / Cloudflare R2 (prod)          |
| Реальное время | WebSocket-хаб (уведомления + личные сообщения)              |
| Reverse proxy  | Caddy 2 (авто-TLS)                                          |
| Деплой         | Yandex Cloud VM по SSH (GitHub Actions) + Cloudflare DNS/CDN |

## Функционал

- Регистрация, логин (email или @username), «запомнить меня», выход.
- Подтверждение e-mail и сброс пароля письмами (одноразовые хэшированные токены).
- Профили: аватар, шапка, био, локация, сайт; счётчики; редактирование.
- Yap'ы: текст (280) + до 4 картинок, ответы (треды), репосты, цитаты, удаление.
- Ленты: «Для вас» (глобальная) и «Подписки» (home), таймлайн профиля
  (yaps / ответы / медиа / лайки), курсорная пагинация.
- Лайки, репосты, закладки, подписки — с оптимистичным UI.
- Поиск по людям и yap'ам, хэштеги (`#тег`), упоминания (`@user`).
- Уведомления (лайк, подписка, ответ, репост, цитата, упоминание) в реальном времени.
- Личные сообщения в реальном времени (WebSocket).
- «Кого читать», русский/английский интерфейс.

## Структура

```
backend/   # Go API (cmd/api, internal/*, db/migrations)
frontend/  # Next.js 14 (src/app/[locale], src/components, src/lib)
deploy/    # прод docker-compose, Caddyfile, .env.prod.example
.github/   # CI и деплой workflows
```

## Быстрый старт (локально)

Нужен только Docker и Docker Compose.

```bash
cp .env.example .env
docker compose up --build
```

Когда всё поднимется:

- Фронтенд (RU):   <http://localhost:3000/ru>
- Фронтенд (EN):   <http://localhost:3000/en>
- API health:      <http://localhost:8080/healthz>
- Mailpit (письма): <http://localhost:8025>
- MinIO консоль:   <http://localhost:9001> (minioadmin / minioadmin)
- Postgres:        `localhost:5432`

Поскольку локально письма уходят в Mailpit, ссылку для подтверждения e-mail
после регистрации открой в его UI (<http://localhost:8025>).

Сброс БД (удаляет том):

```bash
docker compose down -v
```

## Разработка по отдельности

Бэкенд (нужен запущенный Postgres):

```bash
cd backend
export DATABASE_URL="postgres://yapper:yapper@localhost:5432/yapper?sslmode=disable"
make migrate-up   # требует goose: go install github.com/pressly/goose/v3/cmd/goose@latest
make run
```

Фронтенд:

```bash
cd frontend
npm install
NEXT_PUBLIC_API_URL=http://localhost:8080 npm run dev
```

## Деплой (Yandex Cloud + Cloudflare)

1. На VM: установить Docker, склонировать репозиторий в `~/yapper`.
2. `cp deploy/.env.prod.example deploy/.env.prod`, заполнить (домен, Postgres,
   Resend SMTP, Cloudflare R2), `chmod 600 deploy/.env.prod`.
3. В Cloudflare направить домен на IP VM (orange cloud), для медиа — публичный
   R2-домен в `S3_PUBLIC_BASE_URL`.
4. Настроить секреты GitHub: `YC_VM_HOST`, `YC_VM_USER`, `YC_VM_SSH_KEY`,
   `YC_GITHUB_DEPLOY_KEY`, опционально `ADMIN_EMAILS`.
5. Push в `main` запускает `deploy-yc.yml`, который собирает и поднимает
   `docker-compose.prod.yml` (Caddy получает TLS автоматически).

## CI

`.github/workflows/ci.yml` на каждый push/PR: бэкенд — `go vet`, `go test -race`,
`go build`; фронтенд — `lint`, `typecheck`, `build`; затем сборка Docker-образов.
