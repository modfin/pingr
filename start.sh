#!/bin/bash

cd ui
yarn install
bsb -make-world ### TODO maybe add -w & to have the buckle enginen watch and recompile as well
parcel watch index.html &
cd ..

fswatch -config /fsw.yml