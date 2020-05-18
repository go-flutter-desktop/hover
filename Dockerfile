FROM snapcore/snapcraft AS snapcraft
# Using multi-stage dockerfile to obtain snapcraft binary

FROM ubuntu:bionic AS flutterbuilder
RUN apt-get update \
    && apt-get install -y \
        git curl unzip
# Install Flutter from the beta channel
RUN git clone --single-branch --depth=1 --branch beta https://github.com/flutter/flutter /opt/flutter 2>&1 \
    && /opt/flutter/bin/flutter doctor -v

FROM ubuntu:bionic AS xarbuilder
RUN apt-get update \
	&& apt-get install -y \
		git libssl1.0-dev libxml2-dev make g++ autoconf
RUN git clone --single-branch --depth=1 --branch xar-1.6.1 https://github.com/mackyle/xar 2>&1 \
	&& cd xar/xar \
	&& ./autogen.sh --noconfigure \
	&& ./configure 2>&1 \
	&& make 2>&1 \
	&& make install 2>&1

FROM ubuntu:bionic AS bomutilsbuilder
RUN apt-get update \
	&& apt-get install -y \
	    git make g++
RUN git clone --single-branch --depth=1 --branch 0.2 https://github.com/hogliux/bomutils 2>&1 \
	&& cd bomutils \
	&& make 2>&1 \
	&& make install 2>&1

# Fixed using https://github.com/AppImage/AppImageKit/issues/828
FROM ubuntu:bionic as appimagebuilder
RUN apt-get update \
	&& apt-get install -y \
	    wget
RUN cd /opt \
	&& wget -c https://github.com/$(wget -q https://github.com/probonopd/go-appimage/releases -O - | grep "appimagetool-.*-x86_64.AppImage" | head -n 1 | cut -d '"' -f 2) \
	&& mv appimagetool-*-x86_64.AppImage appimagetool-x86_64.AppImage \
	&& chmod a+x appimagetool-x86_64.AppImage \
	&& sed 's|AI\x02|\x00\x00\x00|g' -i appimagetool-x86_64.AppImage \
	&& ./appimagetool-x86_64.AppImage --appimage-extract \
	&& mv squashfs-root appimagetool

FROM dockercore/golang-cross:1.13.11 AS hover

# Install dependencies via apt
RUN apt-get update \
	&& apt-get install -y \
	    # dependencies for compiling linux
		libgl1-mesa-dev xorg-dev \
		# dependencies for darwin-bundle
		icnsutils \
		# dependencies for darwin-dmg
		genisoimage \
		# dependencies for darwin-pkg
		cpio git \
		# dependencies for linux-rpm
		rpm \
		# dependencies for windows-msi
		wixl imagemagick \
	&& rm -rf /var/lib/apt/lists/*

COPY --from=snapcraft /snap /snap
ENV PATH="/snap/bin:$PATH"
ENV SNAP="/snap/snapcraft/current"
ENV SNAP_NAME="snapcraft"
ENV SNAP_ARCH="amd64"

COPY --from=xarbuilder /usr/local/bin/xar /usr/local/bin/xar
COPY --from=xarbuilder /usr/local/lib/libxar.so.1 /usr/local/lib/libxar.so.1
COPY --from=xarbuilder /usr/lib/x86_64-linux-gnu/libcrypto.so.1.0.0 /usr/lib/x86_64-linux-gnu/libcrypto.so.1.0.0

COPY --from=bomutilsbuilder /usr/bin/mkbom /usr/bin/mkbom

COPY --from=appimagebuilder /opt/appimagetool /opt/appimagetool
ENV PATH=/opt/appimagetool/usr/bin:$PATH

# TODO: Add pacman pkg packaging

COPY --from=flutterbuilder /opt/flutter /opt/flutter
RUN ln -sf /opt/flutter/bin/flutter /usr/bin/flutter

# Build hover
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./... 2>&1
RUN go install -v ./... 2>&1

COPY docker/hover-safe.sh /usr/local/bin/hover-safe.sh

WORKDIR /app
