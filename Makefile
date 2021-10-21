.PHONY: build

build:
	go build -o sbin/curveadm $(PWD)/cmd/curveadm/main.go
