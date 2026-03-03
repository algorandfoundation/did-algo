FROM zhield/shell:stable

EXPOSE 9091/tcp

VOLUME ["/etc/algoid"]

COPY algoid /usr/bin/algoid
ENTRYPOINT ["/usr/bin/algoid"]
