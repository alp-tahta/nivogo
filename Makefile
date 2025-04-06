up:
	docker compose up -d

down:
	docker compose down -v

rip:
	docker system prune -a

sal:
	docker compose build --no-cache oms inventory
	docker compose up --no-deps oms inventory

.PHONY: up down rip sal