# vim:set ft=dockerfile:

ARG PG_VERSION=13

FROM postgres:${PG_VERSION}


COPY ./bin/postgresw  /usr/local/sbin/postgresw

ENTRYPOINT ["/usr/local/sbin/postgresw"]