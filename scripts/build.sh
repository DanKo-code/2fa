#!/bin/sh
set -x
export PATH="$PATH:$(go env GOPATH)/bin"
go mod download
go build -o ./build/main ./cmd/main.go #Здесь указать расположение файла main.go