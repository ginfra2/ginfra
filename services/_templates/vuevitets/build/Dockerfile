FROM bitnami/minideb:bookworm

# Overhead
COPY build/container_entrypoint.sh /
RUN chmod 755 /container_entrypoint.sh
COPY build/provision.sh /
RUN chmod 755 /provision.sh
COPY build/custom_provision.sh /
RUN chmod 755 /custom_provision.sh
COPY build/debugging.sh /
RUN chmod 755 debugging.sh

# Server stuff
COPY build/UserDocs.md /UserDocs.md
COPY build/ginfra_service.exe /service/ginfra_service.exe
COPY specifics.json /specifics.json

# Provisioning
RUN /provision.sh
RUN rm -f /provision.sh
RUN /custom_provision.sh
RUN rm -f /custom_provision.sh

# Post provision
COPY /dist/ /usr/share/nginx/html/
COPY build/nginx.conf /etc/nginx/nginx.conf

ENTRYPOINT [ "/container_entrypoint.sh" ]

