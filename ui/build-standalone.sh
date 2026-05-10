#!/usr/bin/env bash
# Like @rancher/shell build-pkg but adds --inline-vue so Vue is bundled into
# the output. Required when component.js is loaded via a script tag (Custom UI URL)
# rather than through the Rancher extension loader.

set -e

SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
BASE_DIR="$SCRIPT_DIR"
SHELL_DIR="$BASE_DIR/node_modules/@rancher/shell"
SHELL_DIR=$(cd -P "${SHELL_DIR}" && pwd)
PKGFILE_JS="$SHELL_DIR/scripts/pkgfile.js"

PKG=$1
if [ -z "$PKG" ]; then
  echo "Usage: $0 <package-name>"
  exit 1
fi

if [ ! -d "${BASE_DIR}/pkg/${PKG}" ]; then
  echo "Package '${PKG}' not found in pkg/"
  exit 1
fi

VERSION=$(cd "pkg/$PKG" && node -p -e "require('./package.json').version")
NAME="${PKG}-${VERSION}"
PKG_DIST="${BASE_DIR}/dist-pkg/${NAME}"

echo "Building standalone UI package $PKG"
echo "  Name:    ${NAME}"
echo "  Output:  ${PKG_DIST}"

rm -rf "${PKG_DIST}"
mkdir -p "${PKG_DIST}"

pushd "pkg/${PKG}"

if [ -e ".shell" ]; then
  LINK=$(readlink .shell)
  if [ "${LINK}" != "${SHELL_DIR}" ]; then
    echo ".shell symlink points to wrong location (${LINK}), expected ${SHELL_DIR}"
    popd; exit 1
  fi
else
  ln -s "${SHELL_DIR}" .shell
fi

FILE=index.js
[ -f ./index.ts ] && FILE=index.ts

"${BASE_DIR}/node_modules/.bin/vue-cli-service" build \
  --name "${NAME}" \
  --target lib \
  --inline-vue \
  "${FILE}" \
  --dest "${PKG_DIST}" \
  --formats umd-min \
  --filename "${NAME}"

cp -f ./package.json "${PKG_DIST}/package.json"
node "${PKGFILE_JS}" "${PKG_DIST}/package.json"
rm -rf "${PKG_DIST}"/*.bak
rm -f .shell

popd
