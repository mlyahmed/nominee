FROM golang:1.15.6-buster as builder

COPY . /pgnominee
WORKDIR /pgnominee
ENV GOBIN=/bin/
RUN CGO_ENABLED=0 GOOS=linux go install ./...

FROM postgres:13
WORKDIR /
COPY --from=builder /bin/pgnominee  .

ENTRYPOINT ["/pgnominee"]