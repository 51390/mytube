FROM golang:1.21
ARG WORKDIR=/service

RUN apt update -y
RUN apt install -y postgresql-client

RUN mkdir $WORKDIR
WORKDIR $WORKDIR
COPY go.mod .
COPY go.sum .
RUN go mod download

ADD . .
RUN make clean bin/server

ENTRYPOINT make run
