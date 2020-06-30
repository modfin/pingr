#!/bin/bash

cd ui
yarn install
bsb -make-world ### TODO maybe add -w & to have the buckle enginen watch and recompile as well
parcel watch index.html &
cd ..

## build step of ui
##
# rm -r ./build
# parcel build -d build index.html
# rm ./build/*.js.map
# go-bindata -pkg ui -o ./fs.go ./build

fswatch -config /fsw.yml