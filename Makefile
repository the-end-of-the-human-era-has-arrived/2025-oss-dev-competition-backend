BIN_DIR ?= bin
NAME := mindmap-server

build:
	go build -o ${BIN_DIR}/${NAME} cmd/main.go
