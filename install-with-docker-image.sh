#!/bin/bash
go install . && DOCKER_BUILDKIT=1 docker build . -t goflutter/hover:latest
