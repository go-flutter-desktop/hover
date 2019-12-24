FROM snapcore/snapcraft AS snapcraft
# Using multi-stage dockerfile to obtain snapcraft binary

FROM ubuntu:bionic AS xarbuilder
RUN apt-get update \
	&& apt-get install -y \
		wget tar libssl1.0-dev libxml2-dev make g++
RUN cd /tmp \
	&& wget https://github.com/downloads/mackyle/xar/xar-1.6.1.tar.gz \
	&& tar -zxvf xar-1.6.1.tar.gz > /dev/null \
	&& mv xar-1.6.1 xar \
	&& cd xar \
	&& ./configure > /dev/null \
	&& make > /dev/null \
	&& make install > /dev/null

FROM goflutter/golang-cross:latest AS hover

# Add dart apt repository
RUN sh -c 'wget -qO- https://dl-ssl.google.com/linux/linux_signing_key.pub | apt-key add -' \
	&& sh -c 'wget -qO- https://storage.googleapis.com/download.dartlang.org/linux/debian/dart_stable.list > /etc/apt/sources.list.d/dart_stable.list'

# Install dependencies via apt
RUN apt-get update \
	&& apt-get upgrade -y \
	&& apt-get install -y \
		# why is gnupg-agent needed? Can this be removed??
		gnupg-agent \
		# for what command are these dependencies??
		libgl1-mesa-dev xorg-dev \
		# dependencies for flutter
		dart unzip \
		# dependencies for darwin-bundle
		icnsutils \
		# dependencies for darwin-dmg
		genisoimage \
		# dependencies for darwin-pkg
		cpio git \
		# dependencies for linux-appimage
		libglib2.0-0 curl file \
		# dependencies for linux-rpm
		rpm \
		# dependencies for linux-snap
		locales \
		# dependencies for windows-msi
		wixl imagemagick \
	&& rm -rf /var/lib/apt/lists/* \
	&& ln -sf /usr/lib/dart/bin/pub /usr/bin/pub

# Install darwin-pkg dependencies
# TODO: make bomutils in a separate stage, copy binaries/libs, like xar.
RUN cd /tmp \
	&& git clone https://github.com/hogliux/bomutils \
	&& cd bomutils \
	&& make > /dev/null \
	&& make install > /dev/null
COPY --from=xarbuilder /usr/local/bin/xar /usr/local/bin/xar
COPY --from=xarbuilder /usr/local/lib/libxar.so.1 /usr/local/lib/libxar.so.1
COPY --from=xarbuilder /usr/lib/x86_64-linux-gnu/libcrypto.so.1.0.0 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.0.0

# Install linux-appimage dependencies
RUN cd /opt \
	&& curl -LO https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage \
	&& chmod a+x appimagetool-x86_64.AppImage \
	&& ./appimagetool-x86_64.AppImage --appimage-extract \
	&& mv squashfs-root appimagetool \
	&& rm appimagetool-x86_64.AppImage
ENV PATH=/opt/appimagetool/usr/bin:$PATH

# Install linux-snap dependencies (based on https://hub.docker.com/r/snapcore/snapcraft/dockerfile)
COPY --from=snapcraft /snap/core /snap/core
COPY --from=snapcraft /snap/snapcraft /snap/snapcraft
COPY --from=snapcraft /snap/bin/snapcraft /snap/bin/snapcraft
# RUN locale-gen en_US.UTF-8 # TODO: remove locales from apt install above
# ENV LANG="en_US.UTF-8"
# ENV LANGUAGE="en_US:en"
# ENV LC_ALL="en_US.UTF-8"
ENV PATH="/snap/bin:$PATH"
ENV SNAP="/snap/snapcraft/current"
ENV SNAP_NAME="snapcraft"
ENV SNAP_ARCH="amd64"
# RUN dpkg-reconfigure locales

# Install Flutter from the beta channel
RUN git clone --single-branch --depth=1 --branch beta https://github.com/flutter/flutter /opt/flutter \
	&& ln -sf /opt/flutter/bin/flutter /usr/bin/flutter \
	&& flutter doctor -v \
	&& flutter config --enable-web

# Build hover
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

COPY docker/hover-safe.sh /usr/local/bin/hover-safe.sh

WORKDIR /app
