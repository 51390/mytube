FROM golang:1.21
ARG WORKDIR=/service

RUN apt update -y
RUN apt install -y postgresql-client

ADD . $WORKDIR
WORKDIR $WORKDIR
RUN make clean bin/server

ENTRYPOINT make run
