.PHONY: modules

modules:
	go build -o ./bin/modules ./cmd/modules/main.go
