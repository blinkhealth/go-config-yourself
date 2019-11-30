FROM techknowlogick/xgo:go-1.12.x
# see https://github.com/techknowlogick/xgo
# Used by go-config-yourself to cross compile to MacOS and Linux from CircleCI

ENV GPGME_VERSION "1.13.0"
ENV GPGERR_VERSION "1.36"
ENV ASSUAN_VERSION "2.5.3"
ENV UPX_VERSION "3.94"

# Get the dependencies and libgpgme
# ADD is broken in docker < 17.06.1
RUN wget --quiet --directory-prefix /build-libs \
  https://github.com/upx/upx/releases/download/v${UPX_VERSION}/upx-${UPX_VERSION}-amd64_linux.tar.xz \
  https://gnupg.org/ftp/gcrypt/libgpg-error/libgpg-error-${GPGERR_VERSION}.tar.bz2 \
  https://gnupg.org/ftp/gcrypt/libassuan/libassuan-${ASSUAN_VERSION}.tar.bz2 \
  https://gnupg.org/ftp/gcrypt/gpgme/gpgme-${GPGME_VERSION}.tar.bz2
RUN tar xf /build-libs/upx-${UPX_VERSION}-amd64_linux.tar.xz \
  && cp upx-${UPX_VERSION}-amd64_linux/upx /usr/local/bin \
  && rm -rf upx-${UPX_VERSION}-amd64_linux /build-libs/upx-${UPX_VERSION}-amd64_linux.tar.xz

# Extract them to versionless directories
RUN mkdir /build-libs/libgpg-error \
  && tar xf /build-libs/libgpg-error-${GPGERR_VERSION}.tar.bz2 -C /build-libs/libgpg-error --strip-components 1
RUN mkdir /build-libs/libassuan \
  && tar xf /build-libs/libassuan-${ASSUAN_VERSION}.tar.bz2 -C /build-libs/libassuan --strip-components 1
RUN mkdir /build-libs/gpgme \
  && tar xf /build-libs/gpgme-${GPGME_VERSION}.tar.bz2 -C /build-libs/gpgme --strip-components 1
# Install it all
COPY gpgme.sh /build-libs
RUN /build-libs/gpgme.sh
# clean up
RUN rm -rf /build-libs
# specify correct prefix for osx cgo builds
RUN sed -i 's|GOOS=darwin GOARCH=amd64|GOOS=darwin GOARCH=amd64 CGO_CFLAGS="-I/usr/local/osx-ndk-x86/SDK/MacOSX10.11.sdk/usr/include" CGO_LDFLAGS="-L/usr/local/osx-ndk-x86/SDK/MacOSX10.11.sdk/usr/lib"|' /build.sh
