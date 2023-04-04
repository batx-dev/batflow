ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONE: all
all:
	@echo "make server, worker, or starter"

.PHONY: starter
starter:
	go run ./starter/...

.PHONY: worker
worker:
	go run ./worker/... 

.PHONY: server
server:
	temporal server start-dev