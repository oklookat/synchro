#!/bin/sh

mkdir -p ../dist
cd ../commander
gomobile bind -target android -androidapi 24 -o ../dist/core.aar
