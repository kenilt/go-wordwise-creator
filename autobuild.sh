#!/bin/sh

# update dependencies
go mod download

# build for MacOS
GOOS=darwin GOARCH=amd64 go build

# create zip file for MacOS
[ -f go-wordwise-creator-macos ] && rm -r go-wordwise-creator-macos
mkdir go-wordwise-creator-macos
cp -R resources go-wordwise-creator-macos
mv go-wordwise-creator go-wordwise-creator-macos
zip -r go-wordwise-creator-macos.zip go-wordwise-creator-macos/
rm -r go-wordwise-creator-macos

# build for Windows
GOOS=windows GOARCH=amd64 go build

# create zip file for Windows
[ -f go-wordwise-creator-windows ] && rm -r go-wordwise-creator-windows
mkdir go-wordwise-creator-windows
cp -R resources go-wordwise-creator-windows
mv go-wordwise-creator.exe go-wordwise-creator-windows
zip -r go-wordwise-creator-windows.zip go-wordwise-creator-windows/
rm -r go-wordwise-creator-windows

# Move build to /build-outputs
[ -f build-outputs ] && rm -r build-outputs
mkdir build-outputs
mv go-wordwise-creator-macos.zip build-outputs
mv go-wordwise-creator-windows.zip build-outputs
