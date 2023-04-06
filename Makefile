ifneq (,$(wildcard ./.env))
    include .env
    export
endif

.PHONE: all
all:
	@echo "make server, worker, or apiserver"

.PHONY: apiserver
apiserver:
	go run ./cmd/batflow/... apiserver 

.PHONY: worker
worker:
	go run ./cmd/batflow/... worker 

.PHONY: server
server:
	temporal server start-dev