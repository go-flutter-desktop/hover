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
	    curl
RUN cd /opt \
	&& curl -LO https://github.com/AppImage/AppImageKit/releases/download/continuous/appimagetool-x86_64.AppImage \
	&& chmod a+x appimagetool-x86_64.AppImage \
	&& sed 's|AI\x02|\x00\x00\x00|g' -i appimagetool-x86_64.AppImage \
	&& ./appimagetool-x86_64.AppImage --appimage-extract \
	&& mv squashfs-root appimagetool

# bionic ships with a too old meson version
FROM ubuntu:focal AS pacmanbuilder
ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update \
    && apt-get install -y \
        git meson python3 python3-pip python3-setuptools python3-wheel ninja-build gcc pkg-config m4 libarchive-dev libssl-dev
RUN cd /tmp \
    && git clone https://git.archlinux.org/pacman.git --depth=1 --branch=v5.2.2 2>&1  \
    && cd pacman \
    && meson setup builddir \
    && meson install -C builddir

FROM dockercore/golang-cross:1.13.15 AS hover

# Install dependencies via apt
RUN apt-get update \
	&& apt-get install -y \
	    # dependencies for compiling linux
		libgl1-mesa-dev xorg-dev \
		# dependencies for compiling windows
		wine \
		# dependencies for darwin-dmg
		genisoimage \
		# dependencies for darwin-pkg
		cpio git \
		# dependencies for linux-rpm
		rpm \
		# dependencies for linux-pkg
		fakeroot bsdtar \
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

COPY --from=pacmanbuilder /usr/bin/makepkg /usr/bin/makepkg
COPY --from=pacmanbuilder /usr/bin/pacman /usr/bin/pacman
COPY --from=pacmanbuilder /etc/makepkg.conf /etc/makepkg.conf
COPY --from=pacmanbuilder /etc/pacman.conf /etc/pacman.conf
COPY --from=pacmanbuilder /usr/share/makepkg /usr/share/makepkg
COPY --from=pacmanbuilder /usr/share/pacman /usr/share/pacman
COPY --from=pacmanbuilder /var/lib/pacman /var/lib/pacman
COPY --from=pacmanbuilder /usr/lib/x86_64-linux-gnu/libalpm.so.12 /usr/lib/x86_64-linux-gnu/libalpm.so.12
RUN ln -sf /bin/bash /usr/bin/bash
RUN sed -i "s/OPTIONS=(strip /OPTIONS=(/g" /etc/makepkg.conf
RUN sed -i "s/#XferCommand/XferCommand/g" /etc/pacman.conf
# This makes makepkg believe we are not root. Bypassing the root check is ok, because we are in a container
ENV EUID=1

# Create symlink for darwin-dmg
RUN ln -s $(which genisoimage) /usr/bin/mkisofs

COPY --from=flutterbuilder /opt/flutter /opt/flutter
RUN ln -sf /opt/flutter/bin/flutter /usr/bin/flutter

# Build hover
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./... 2>&1
RUN go install -v ./... 2>&1

COPY docker/hover-safe.sh /usr/local/bin/hover-safe.sh

WORKDIR /app
