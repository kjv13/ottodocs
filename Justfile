DATE := `date +"%Y-%m-%d_%H:%M:%S"`
GIT_COMMIT := `git rev-parse HEAD`
VERSION_TAG := `git describe --tags --abbrev=0`
LD_FLAGS := "-X github.com/chand1012/ottodocs/cmd.buildDate=" + DATE + " -X github.com/chand1012/ottodocs/cmd.commitHash=" + GIT_COMMIT + " -X github.com/chand1012/ottodocs/cmd.tag=" + VERSION_TAG

default:
  just --list --unsorted

tidy:
  go mod tidy

build:
  go build -ldflags "{{LD_FLAGS}}" -v -o bin/otto

clean:
  rm -rf bin dist

run *commands:
  go run main.go {{commands}}

cobra-docs:
  rm docs/*.md
  go run docs/gen_docs.go

install: build
  rm -rf $GOPATH/bin/otto
  cp bin/otto $GOPATH/bin

crossbuild:
  #!/bin/bash

  # Set the name of the output binary and Go package
  BINARY_NAME="otto"
  GO_PACKAGE="github.com/chand1012/ottodocs"

  mkdir -p dist

  # Build for M1 Mac (Apple Silicon)
  echo "Building for M1 Mac (Apple Silicon)"
  env GOOS=darwin GOARCH=arm64 go build -ldflags "{{LD_FLAGS}}" -o "${BINARY_NAME}" "${GO_PACKAGE}"
  zip "${BINARY_NAME}_darwin_arm64.zip" "${BINARY_NAME}"
  rm "${BINARY_NAME}"
  mv "${BINARY_NAME}_darwin_arm64.zip" dist/

  # Build for AMD64 Mac (Intel)
  echo "Building for AMD64 Mac (Intel)"
  env GOOS=darwin GOARCH=amd64 go build -ldflags "{{LD_FLAGS}}" -o "${BINARY_NAME}" "${GO_PACKAGE}"
  zip "${BINARY_NAME}_darwin_amd64.zip" "${BINARY_NAME}"
  rm "${BINARY_NAME}"
  mv "${BINARY_NAME}_darwin_amd64.zip" dist/

  # Build for AMD64 Windows
  echo "Building for AMD64 Windows"
  env GOOS=windows GOARCH=amd64 go build -ldflags "{{LD_FLAGS}}" -o "${BINARY_NAME}.exe" "${GO_PACKAGE}"
  zip "${BINARY_NAME}_windows_amd64.zip" "${BINARY_NAME}.exe"
  rm "${BINARY_NAME}.exe"
  mv "${BINARY_NAME}_windows_amd64.zip" dist/

  # Build for AMD64 Linux
  echo "Building for AMD64 Linux"
  env GOOS=linux GOARCH=amd64 go build -ldflags "{{LD_FLAGS}}" -o "${BINARY_NAME}" "${GO_PACKAGE}"
  tar czvf "${BINARY_NAME}_linux_amd64.tar.gz" "${BINARY_NAME}"
  rm "${BINARY_NAME}"
  mv "${BINARY_NAME}_linux_amd64.tar.gz" dist/

  # Build for ARM64 Linux
  echo "Building for ARM64 Linux"
  env GOOS=linux GOARCH=arm64 go build -ldflags "{{LD_FLAGS}}" -o "${BINARY_NAME}" "${GO_PACKAGE}"
  tar czvf "${BINARY_NAME}_linux_arm64.tar.gz" "${BINARY_NAME}"
  rm "${BINARY_NAME}"
  mv "${BINARY_NAME}_linux_arm64.tar.gz" dist/
