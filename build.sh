#!/bin/bash

NAME=pingrd
IMAGE_NAME=modfin/${NAME}

docker build -f ./Dockerfile.build -t ${IMAGE_NAME}:latest -t ${IMAGE_NAME}:0.0.1 .

docker push modfin/${NAME}:latest
docker push modfin/${NAME}:0.0.1

echo -- start by --
echo docker run -i -p 8080:8080 ${IMAGE_NAME}
