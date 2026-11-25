FROM bitnami/minideb:bookworm

# Overhead
COPY build/container_entrypoint.sh /
COPY build/provision.sh /
RUN chmod 755 /provision.sh
COPY build/custom_provision.sh /
RUN chmod 755 /custom_provision.sh
COPY build/custom_entrypoint.sh /custom_entrypoint.sh
COPY build/debugging.sh /

# Server stuff
COPY test/* /service/test/
COPY build/ginfra_service.exe /service/ginfra_service.exe
COPY doc/UserDocs.md /UserDocs.md
COPY specifics.json /specifics.json

ADD build/files /


COPY ginfra.js package.json package-lock.json server.js  /service/
COPY bin /service/bin/
COPY public /service/public/

# Provisioning
RUN /provision.sh
RUN rm -f /provision.sh
RUN /custom_provision.sh
RUN rm -f /custom_provision.sh

ENTRYPOINT [ "/container_entrypoint.sh" ]
