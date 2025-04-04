up:
	docker compose up -d

down:
	docker compose down -v

rip:
	docker system prune -a

.PHONY: up down rip