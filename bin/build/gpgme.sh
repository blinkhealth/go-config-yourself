#!/usr/bin/env bash
# Builds gpgme and dependencies for linux and macOS

set -o errexit

BUILD_ROOT="/build-libs"
targets=(linux arm darwin)
gpgme="$BUILD_ROOT/gpgme"
libs=("$BUILD_ROOT/libgpg-error" "$BUILD_ROOT/libassuan")
for target in "${targets[@]}"; do
  mkdir -p "${BUILD_ROOT}/$target/libs"
  # copy libs to their target's dir, and strip the version
  cp -r "$gpgme" "${BUILD_ROOT}/$target/${gpgme##*/}"
  for lib in "${libs[@]}"; do
    cp -r "$lib" "${BUILD_ROOT}/$target/libs/${lib##*/}"
  done
done

# Quiet down
export OSXCROSS_NO_INCLUDE_PATH_WARNINGS=1
buildEnv() {
  case $1 in
    darwin)
      HOST=x86_64-apple-darwin
      # GPGME requires host to be different, but HOST to be set
      # I don't quite understand why, but there's only that directory in the source
      export GPGME_HOST='x86_64-apple-darwin15'
      # use special compilers
      export CC=o64-clang
      export CXX=o64-clang++
      # use special libtool and friends
      export LIBTOOL="x86_64-apple-darwin15-libtool"
      export AR="x86_64-apple-darwin15-ar"
      export RANLIB="x86_64-apple-darwin15-ranlib"
      # This is where the compiled targets go
      export PREFIX=/usr/local/osx-ndk-x86/SDK/MacOSX10.11.sdk/usr
      export PATH="/usr/local/osx-ndk-x86/bin:$PATH"
      ;;
    linux)
      HOST=x86_64-linux
      export GPGME_HOST="$HOST"
      export PREFIX=/usr/local
      ;;
    arm)
      HOST=arm-linux-gnueabihf
      export GPGME_HOST="$HOST"
      export CC=arm-linux-gnueabihf-gcc-5
      export CXX=arm-linux-gnueabihf-g++-5
      export PREFIX=/usr/arm-linux-gnueabihf
      export CFLAGS="-march=armv7-a -fPIC"
      export CXXFLAGS="-march=armv7-a -fPIC"
      export PKG_CONFIG_PATH=/usr/arm-linux-gnueabihf/lib/pkgconfig
      export AR="arm-linux-gnueabihf-ar"
      export LD="arm-linux-gnueabihf-ld"
      export CPP="arm-linux-gnueabihf-cpp-5"
      export STRIP="arm-linux-gnueabihf-strip"
      export RANLIB="arm-linux-gnueabihf-ranlib"
      ;;
    default)
      >&2 echo "Unknown target $1"
  esac
  export HOST
}

for target in "${targets[@]}"; do
  for lib in "${libs[@]}"; do
    libName="${lib##*/}"
    echo "Building $libName for $target"
    touch "$lib.log"
    if ! (
      # subshell to get a clean environment every time
      buildEnv "$target"
      cd "$lib"
      make clean
      ./configure \
        --host="$HOST" \
        --prefix="$PREFIX" \
        --with-gpg-error-prefix="$PREFIX" \
        --disable-doc \
        --disable-tests \
        --disable-languages \
        --disable-dependency-tracking \
        --enable-static --disable-shared
      make
      make install
    ) > "$lib.log" 2>&1; then
      echo "Failed building $libName for $target"
      cat "$lib.log"
      exit 2
    fi
  done

  echo "Building gpgme for $target"
  touch "$gpgme.log"
  if ! (
    buildEnv "$target"
    cd "$gpgme"
    make clean
    ./configure \
      --host="${GPGME_HOST}" \
      --prefix="$PREFIX" \
      --with-libgpg-error-prefix="$PREFIX" \
      --with-libassuan-prefix="$PREFIX" \
      --disable-shared --enable-static \
      --disable-gpgconf-test \
      --disable-languages \
      --disable-gpgsm-test \
      --disable-g13-test \
      --disable-gpg-test
    make
    make install
  ) > "$gpgme.log" 2>&1 ; then
    echo "failed building gpgme for $target"
    cat "$gpgme.log"
    exit 2
  fi
done
