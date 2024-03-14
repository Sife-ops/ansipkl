.PHONY: ansipkl modules

ansipkl:
	go build

modules:
	go build -o ./bin/modules ./cmd/modules/main.go
