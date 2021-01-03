# vim:set ft=dockerfile:

ARG HAPROXY_VERSION=2.3
FROM haproxy:${HAPROXY_VERSION}

COPY ./bin/haproxyw  /usr/local/sbin/haproxyw
COPY ./docker/haproxyw/haproxy.cfg /usr/local/etc/haproxy/haproxy.cfg

WORKDIR /usr/local/sbin

ENTRYPOINT ["/usr/local/sbin/haproxyw"]