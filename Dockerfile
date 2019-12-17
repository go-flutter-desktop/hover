FROM golang:1.13-buster

RUN apt-get update && \
	apt-get install unzip imagemagick apt-transport-https ca-certificates curl wget gnupg-agent software-properties-common -y && \
	sh -c 'wget -qO- https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -' && \
	sh -c 'wget -qO- https://storage.googleapis.com/download.dartlang.org/linux/debian/dart_stable.list > /etc/apt/sources.list.d/dart_stable.list' && \
	curl -fsSL https://download.docker.com/linux/debian/gpg | apt-key add - && \
	add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/debian buster stable" && \
	apt-get update && \
	apt-get install dart docker-ce docker-ce-cli containerd.io -y && \
	ln -sf /usr/lib/dart/bin/pub /usr/bin/pub

RUN git clone --single-branch --branch beta https://github.com/flutter/flutter /opt/flutter && \
	ln -sf /opt/flutter/bin/flutter /usr/bin/flutter && \
	flutter doctor -v && \
	flutter config --enable-web

WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...
