#!/usr/bin/env bash

go build -ldflags "-X github.com/dmitriy-rs/rollercoaster/internal/logger.MODE=PROD" -o rollercoaster ./main.go

sudo mv rollercoaster /usr/local/bin/rollercoaster