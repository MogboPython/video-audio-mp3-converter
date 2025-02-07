.PHONY: dev prod

dev:
	GO_ENV=development go run cmd/server/main.go

prod:
	GO_ENV=production go run cmd/server/main.go

docker-dev:
	docker-compose -f docker-compose.dev.yml up --build

docker-prod:
	docker-compose -f docker-compose.prod.yml up --build 