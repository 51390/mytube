FROM ubuntu

RUN apt update -y; apt install -y postgresql-client

COPY entrypoint.sh /usr/bin/entrypoint.sh

RUN chmod a+x /usr/bin/entrypoint.sh

ENTRYPOINT /usr/bin/entrypoint.sh
