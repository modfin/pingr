#!/bin/bash

NAME=pingr
IMAGE_NAME=${NAME}:latest

docker build -f ./Dockerfile.build -t ${IMAGE_NAME} .

echo -- start by --
echo docker run -i -p 8080:8080 pingr:latest
