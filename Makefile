up:
	docker compose down
	docker builder prune -f
	docker compose up --build -d
