#!/usr/bin/env bash

go build -o rollercoaster ./main.go
sudo mv rollercoaster /usr/local/bin/rollercoaster