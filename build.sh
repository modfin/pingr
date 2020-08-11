#!/bin/bash


NAME=pingrd
IMAGE_NAME=modfin/${NAME}

docker build -f ./Dockerfile.snap -t ${IMAGE_NAME}:snap  .
docker create -ti --name dummy modfin/pingrd:snap bash
docker cp dummy:/pingrd .
docker rm -fv dummy