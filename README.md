# ZARA - Go Web Application

Простое веб-приложение на Go с подключением к базой данных PostgreSQL через GORM, авторизацией по сессиям и системой комментариев.

---

## 🚀 Установка и запуск

1. Клонировать репозиторий:
   git clone https://github.com/DEIN_USERNAME/ZARA.git
   cd ZARA

2. Скачать зависимости:
   go mod download

3. Создать файл `.env`:
   Создай файл `.env` в корневой папке проекта и добавь данные для подключения к PostgreSQL:
   DB_DSN="host=localhost user=DEIN_PG_USER password=DEIN_PG_PASSWORD dbname=meindb port=5432 sslmode=disable"

4. Запустить PostgreSQL:
   Убедись, что служба PostgreSQL запущена и база данных `meindb` создана.

5. Запустить приложение:
   go run main.go

Приложение будет доступно в браузере по адресу: http://localhost:7777