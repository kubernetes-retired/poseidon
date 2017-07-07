#!/bin/bash
cd ../
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w'
cp poseidon deploy/
cd deploy/
docker build -t "poseidon:dev" .
