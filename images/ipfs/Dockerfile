FROM debian:10

ENV DEBIAN_FRONTEND noninteractive

ARG version=0.10.0

# Install node tools
RUN \
  apt update && \
  apt install -y \
    wget \
    libssl-dev \
    ca-certificates \
    fuse \
    unzip && \
  wget https://dist.ipfs.io/go-ipfs/v${version}/go-ipfs_v${version}_linux-amd64.tar.gz && \
  tar -xvzf go-ipfs_v${version}_linux-amd64.tar.gz && \
  ./go-ipfs/install.sh && \
  rm -rf go-ipfs* && \
  mkdir /var/lib/ipfs

COPY run.sh /var/run/ipfs.sh

# This is required to facilitate the process of loading the webui locally.
# Get the latest CID from:
#   https://github.com/ipfs/go-ipfs/blob/master/core/corehttp/webui.go
# Download the latest published version:
#   curl https://ipfs.io/api/v0/get/${CID} | tar -xf -
# Install locally:
#   ipfs add -r --cid-version=1 build
COPY webui-v2.13.0.zip /root/webui.zip

# Node storage
VOLUME /var/lib/ipfs
ENV IPFS_PATH="/var/lib/ipfs"

# Entry path
WORKDIR /var/lib/ipfs

# Swarm TCP, UDP and WebSockets (4002/ws)
EXPOSE 4001 4001/udp 4002

# Daemon API
EXPOSE 5001

# Web Gateway
EXPOSE 8080

# Default binary
ENTRYPOINT ["/var/run/ipfs.sh"]
