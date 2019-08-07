#!/usr/bin/env bash
# Builds gpgme and dependencies for linux and macOS

set -o errexit

BUILD_ROOT="/build-libs"
targets=(linux darwin)
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
  if [[ $1 == "darwin" ]]; then
    HOST=x86_64-apple-darwin
    # GPGME requires host to be different, but HOST to be set
    # I don't quite understand why, but there's only that directory in the source
    export GPGME_HOST='x86_64-apple-darwin15'
    # use special compilers
    export CC=o64-clang
    export CXX=o64-clang++
    # This is where the compiled targets go
    export PREFIX=/usr/local/osx-ndk-x86/SDK/MacOSX10.11.sdk/usr
  else
    HOST=x86_64-linux
    export GPGME_HOST="$HOST"
    export PREFIX=/usr/local
    # When building for linux, we want static libraries only, macOS are static by default i think?
    export LIB_FLAGS="--disable-shared --enable-static"
  fi
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
        --disable-doc \
        --disable-tests \
        --disable-languages $LIB_FLAGS
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
