FROM ghcr.io/bryk-io/shell:0.2.0

EXPOSE 9090/tcp

VOLUME ["/etc/algoid"]

COPY algoid /usr/bin/algoid
ENTRYPOINT ["/usr/bin/algoid"]
