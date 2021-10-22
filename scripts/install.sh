#!/bin/sh

# Utilities
uname_arch() {
  arch=$(uname -m)
  case $arch in
    x86_64) arch="amd64" ;;
    x86) arch="386" ;;
    i686) arch="386" ;;
    i386) arch="386" ;;
    aarch64) arch="arm64" ;;
  esac
  echo ${arch}
}

uname_os() {
  os=$(uname -s | tr '[:upper:]' '[:lower:]')
  echo "$os"
}

# Check requirements
if ! [ -x "$(command -v sudo)" ]; then
  echo "this installer requires 'sudo' to be installed"
  exit
fi

if ! [ -x "$(command -v curl)" ]; then
  echo "this installer requires 'curl' to be installed"
  exit
fi

# Setup
GITHUB_OWNER="algorandfoundation"
GITHUB_REPO="did-algo"
TAG="latest"
INSTALL_DIR="/usr/local/bin/"
ARCH=$(uname_arch)
OS=$(uname_os)
GITHUB_RELEASES_PAGE=https://github.com/${GITHUB_OWNER}/${GITHUB_REPO}/releases
VERSION=$(curl $GITHUB_RELEASES_PAGE/$TAG -sL -H 'Accept:application/json' | tr -s '\n' ' ' | sed 's/.*"tag_name":"//' | sed 's/".*//' | tr -d v)
NAME=${GITHUB_REPO}_${VERSION}_${OS}_${ARCH}
TARBALL=${NAME}.tar.gz
TARBALL_URL=${GITHUB_RELEASES_PAGE}/download/v${VERSION}/${TARBALL}

# Install package
echo "Downloading latest $GITHUB_REPO binary from $TARBALL_URL"
tmpfolder=$(mktemp -d)
$(curl $TARBALL_URL -sL -o $tmpfolder/$TARBALL)

if [ ! -f $tmpfolder/$TARBALL ]; then
  echo "Can not download. Exiting..."
  exit 14
fi
cd ${tmpfolder} && tar --no-same-owner -xzf "$tmpfolder/$TARBALL"

if [ ! -f $tmpfolder/$GITHUB_REPO ]; then
  echo "Can not find $GITHUB_REPO. Exiting..."
  exit 15
fi

binary=$tmpfolder/$GITHUB_REPO
echo "Installing $GITHUB_REPO to $INSTALL_DIR"
sudo install "$binary" $INSTALL_DIR
echo "Installed $GITHUB_REPO to $INSTALL_DIR"

# Clean-up
rm -rf "${tmpdir}"
