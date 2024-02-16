#!/bin/sh

mkdir -p ../dist
cd ..
go build -o "dist/core.so" -buildmode=c-shared
