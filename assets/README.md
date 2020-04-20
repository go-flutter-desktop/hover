# assets

This directory contains templates and config files that hover uses to initialize apps and packaging structures. When modifying these assets, you need to update the generated code so that the assets are included in the Go build process.

## Installing rice

Install the rice tool by running `(cd $HOME && GO111MODULE=on go get -u -a github.com/GeertJohan/go.rice/rice)`.

## Updating code

Run `go generate ./...` in the repository to update the generated code.
