FROM python:3.11
ARG WORKDIR=/app

RUN apt update -y
RUN apt install -y postgresql-client

RUN mkdir $WORKDIR
WORKDIR $WORKDIR
COPY requirements.txt .
COPY Makefile .
RUN make init-app

ADD . .
ENTRYPOINT make run-app-dev
