.PHONY: build

build:
	go build -o sbin/curveadm $(PWD)/cmd/curveadm/main.go

debug:
	go build -o ~/.curveadm/bin/curveadm $(PWD)/cmd/curveadm/main.go
