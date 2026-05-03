#!/usr/bin/env bash


# start cluster to do integrate test
go build -o starter ../cmd/starter/starter.go


./starter -config start1.json &
./starter -config start2.json &
./starter -config start3.json &


