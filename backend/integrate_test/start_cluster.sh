#!/usr/bin/env bash

script_dir=$(dirname "$0")
script_name=$(basename "$0")

cd "$script_dir" || exit 1

# start cluster to do integrate test
# go build -o starter ../cmd/starter/starter.go

./starter -config start1.json &
./starter -config start2.json &
./starter -config start3.json &


