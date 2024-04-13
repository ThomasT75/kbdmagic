#!/usr/bin/env bash
#go build -ldflags="-s -w" .

go mod tidy
go build .
