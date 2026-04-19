.PHONY: compose-up compose-down

# Поднять инфраструктуру (базы данных, брокер)
compose-up:
	docker compose up -d

# Опустить инфраструктуру
compose-down:
	docker compose down -v