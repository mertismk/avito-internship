.PHONY: help build up down logs restart clean

help:
	@echo "Доступные команды:"
	@echo "make build - собрать Docker образ"
	@echo "make up - запустить все сервисы"
	@echo "make down - остановить все сервисы"
	@echo "make logs - показать логи приложения"
	@echo "make restart - перезапустить сервисы"
	@echo "make clean - очистить образы"

build:
	docker compose build

up:
	docker compose up -d
	@echo "Сервис запущен на http://localhost:8080"

down:
	docker compose down

logs:
	docker compose logs -f app

restart: down up

clean:
	docker compose down -v