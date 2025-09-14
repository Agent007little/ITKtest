# Используем официальный образ Go как базовый
FROM golang:1.23-alpine as builder

# Устанавливаем рабочую директорию внутри контейнера
WORKDIR /app

# Устанавливаем git (необходим для go install)
RUN apk add --no-cache git

# Копируем исходники приложения в рабочую директорию
COPY . .

# Скачиваем все зависимости
RUN go mod download && go mod verify

# Собираем приложение
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Начинаем новую стадию сборки на основе минимального образа
FROM alpine:latest

# Копируем .env файл (если существует)
COPY config.env .

# Копируем папку migrations
COPY migrations ./migrations

# Добавляем исполняемый файл из первой стадии в корневую директорию контейнера
COPY --from=builder /app/main /main

# Открываем порт 8080
EXPOSE 8080

# Запускаем приложение
CMD ["/main"]