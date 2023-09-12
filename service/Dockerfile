FROM golang:1.21
ARG WORKDIR=/service

ADD . $WORKDIR
WORKDIR $WORKDIR
RUN make clean bin/server

ENTRYPOINT make run
