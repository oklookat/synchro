#!/bin/sh

mkdir -p ../dist
cd ../commander
gomobile bind -target ios -o ../dist/core.framework -bundleid "com.oklookat.synchro"
