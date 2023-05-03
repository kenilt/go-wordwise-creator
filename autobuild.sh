#!/bin/sh

# update dependencies
go mod download

# build for MacOS intel
GOOS=darwin GOARCH=amd64 go build

# create zip file for MacOS intel
rm -rf go-wordwise-creator-macos-amd64
mkdir go-wordwise-creator-macos-amd64
cp -R resources go-wordwise-creator-macos-amd64
mv go-wordwise-creator go-wordwise-creator-macos-amd64
zip -r go-wordwise-creator-macos-amd64.zip go-wordwise-creator-macos-amd64/
rm -r go-wordwise-creator-macos-amd64

# build for MacOS arm
GOOS=darwin GOARCH=arm64 go build

# create zip file for MacOS arm
rm -rf go-wordwise-creator-macos-arm64
mkdir go-wordwise-creator-macos-arm64
cp -R resources go-wordwise-creator-macos-arm64
mv go-wordwise-creator go-wordwise-creator-macos-arm64
zip -r go-wordwise-creator-macos-arm64.zip go-wordwise-creator-macos-arm64/
rm -r go-wordwise-creator-macos-arm64

# build for Windows
GOOS=windows GOARCH=amd64 go build

# create zip file for Windows
rm -rf go-wordwise-creator-windows
mkdir go-wordwise-creator-windows
cp -R resources go-wordwise-creator-windows
mv go-wordwise-creator.exe go-wordwise-creator-windows
zip -r go-wordwise-creator-windows.zip go-wordwise-creator-windows/
rm -r go-wordwise-creator-windows

# Move build to /build-outputs
rm -rf build-outputs
mkdir build-outputs
mv go-wordwise-creator-macos-amd64.zip build-outputs
mv go-wordwise-creator-macos-arm64.zip build-outputs
mv go-wordwise-creator-windows.zip build-outputs
