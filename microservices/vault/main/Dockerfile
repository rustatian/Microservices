FROM iron/go
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD config config
ADD vaultd vaultd
EXPOSE 10000 10000
ENTRYPOINT ./vaultd -consul.addr=0xdev.me