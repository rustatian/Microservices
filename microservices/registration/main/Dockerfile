FROM iron/go
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD regd regd
ADD config config
EXPOSE 10002 10002
EXPOSE 3306
ENTRYPOINT ./regd -consul.addr=0xdev.me