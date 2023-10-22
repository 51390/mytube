FROM python:3.11
ARG WORKDIR=/app

RUN apt update -y
RUN apt install -y postgresql-client

ADD . $WORKDIR
WORKDIR $WORKDIR
RUN make init-app

ENTRYPOINT make run-app-dev
