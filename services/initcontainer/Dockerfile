FROM bitnami/minideb:bookworm

# Overhead
COPY container_entrypoint.sh /
RUN chmod 755 /container_entrypoint.sh

ENTRYPOINT [ "/container_entrypoint.sh" ]

