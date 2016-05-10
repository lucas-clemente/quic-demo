FROM ubuntu:16.04

RUN apt-get update && apt-get -y install python iproute sudo && rm -rf /var/lib/apt/lists/*

EXPOSE 8000 8001 8002 8003 8004 8005 8006 8007 8008 8009
VOLUME /certs

ADD run.sh /run.sh
ADD main /main

CMD ["/run.sh"]
