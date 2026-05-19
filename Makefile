APP_NAME=inventory-service
BACKEND_DIR=backend

GO=go
DOCKER_COMPOSE=docker compose

# =========================
# Development
# =========================

.PHONY: help
help:
	@echo "Available commands:"
	@echo "  make run               - Run HTTP server"
	@echo "  make migration         - Run database migrations"
	@echo "  make test              - Run test"
	@echo "  make daily-report      - Get daily report"

# =========================
# Go Commands
# =========================

.PHONY: run
run:
	cd $(BACKEND_DIR) && $(GO) run ./cmd/server

.PHONY: migration
migration:
	cd $(BACKEND_DIR) && $(GO) run ./cmd/migrations

.PHONY: test
test:
	cd $(BACKEND_DIR) && $(GO) test ./... -v

.PHONY: daily-report
daily-report:
	cd $(BACKEND_DIR) && $(GO) run ./cmd/report

# =========================
# Docker
# =========================

.PHONY: docker-up
docker-up:
	$(DOCKER_COMPOSE) up -d

.PHONY: docker-down
docker-down:
	$(DOCKER_COMPOSE) down

.PHONY: docker-build
docker-build:
	$(DOCKER_COMPOSE) build

.PHONY: docker-logs
docker-logs:
	$(DOCKER_COMPOSE) logs -f

.PHONY: docker-restart
docker-restart:
	$(DOCKER_COMPOSE) restart

# =========================
# Full setup
# =========================

.PHONY: dev
dev: docker-up migration run

.PHONY: reset
reset: docker-down
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) up -d
