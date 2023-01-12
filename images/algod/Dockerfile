FROM debian:stable-slim

ENV DEBIAN_FRONTEND noninteractive

# To build for a different channel try passing:
# --build-arg channel=beta
ARG channel=stable

# Install node tools
RUN \
  apt update && \
  apt install -y curl && \
  mkdir -p /var/lib/algorand/genesis && \
  curl https://raw.githubusercontent.com/algorand/go-algorand-doc/master/downloads/installers/update.sh -O && \
  chmod 544 update.sh && \
  ./update.sh -i -c ${channel} -p /var/lib/algorand/bin -g /var/lib/algorand/ -d /var/lib/algorand/data -n && \
  mv /var/lib/algorand/bin/genesisfiles/* /var/lib/algorand/genesis

# Add executable to global path
ENV PATH="${PATH}:/var/lib/algorand/bin"

# Base settings
COPY system.json /var/lib/algorand/system.json
COPY config.json /var/lib/algorand/config.json
COPY run.sh /var/lib/algorand/run.sh

# Node storage
VOLUME /var/lib/algorand/data
ENV ALGORAND_DATA="/var/lib/algorand/data"

# Entry path
WORKDIR /var/lib/algorand

# RPC port
EXPOSE 8080

# Default binary
ENTRYPOINT ["/var/lib/algorand/run.sh"]
