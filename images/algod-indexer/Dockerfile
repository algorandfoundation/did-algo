FROM debian:10

ENV DEBIAN_FRONTEND noninteractive

ARG version=2.12.4

# Install base tools
RUN \
  apt update && \
  apt install -y ca-certificates wget

# Install indexer tools
RUN \
  wget https://github.com/algorand/indexer/releases/download/${version}/algorand-indexer_${version}_amd64.deb && \
  dpkg -i algorand-indexer_${version}_amd64.deb && \
  rm algorand-indexer_${version}_amd64.deb

# Use a non-privilidged user for security reasons
RUN \
  mkdir /opt/indexer && \
  groupadd -g 10353 indexer && \
  useradd -m -u 10353 -g indexer indexer && \
  chown -R indexer:indexer /opt && \
  ls -lha /opt
USER indexer

# API port
EXPOSE 8980

# Default binary
ENTRYPOINT ["/usr/bin/algorand-indexer"]
