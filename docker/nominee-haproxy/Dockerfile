# vim:set ft=dockerfile:

ARG HAPROXY_VERSION=2.3
FROM haproxy:${HAPROXY_VERSION}

COPY ./bin/nominee-haproxy  /usr/local/sbin/nominee-haproxy
COPY ./docker/nominee-haproxy/haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg

WORKDIR /usr/local/sbin

ENTRYPOINT ["nominee-haproxy"]