.PHONY: starter
starter:
	go run ./starter/...

.PHONY: worker
worker:
	go run ./worker/...

.PHONY: server
server:
	temporal server start-dev