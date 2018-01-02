FROM iron/go
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD authd authd
ADD config config
EXPOSE 10001 10001
EXPOSE 3306
EXPOSE 6379
ENTRYPOINT ./authd -consul.addr=0xdev.me