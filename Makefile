.PHONY: dev prod

# Run in development mode
dev:
	APP_ENV=dev go run ./cmd/api

# Run in production mode
prod:
	APP_ENV=prod go run ./cmd/api

