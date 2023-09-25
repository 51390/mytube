FROM python:3.11
ARG WORKDIR=/app

ADD . $WORKDIR
WORKDIR $WORKDIR
RUN make init-app

ENTRYPOINT make init-db run-app-dev
