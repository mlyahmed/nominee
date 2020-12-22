# vim:set ft=dockerfile:

ARG PG_VERSION=13

FROM postgres:${PG_VERSION}


COPY ./bin/nominee-postgres  /usr/local/sbin/nominee-postgres

ENTRYPOINT ["nominee-postgres"]