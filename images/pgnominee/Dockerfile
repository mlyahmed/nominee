# vim:set ft=dockerfile:

ARG PG_VERSION=13

FROM postgres:${PG_VERSION}
WORKDIR /
COPY ./bin/pgnominee  .

ENTRYPOINT ["/pgnominee"]