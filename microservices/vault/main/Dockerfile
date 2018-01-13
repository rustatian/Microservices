FROM golang
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD /microservices/vault vault
ADD /config vault/main/config
EXPOSE 10000 10000
WORKDIR vault/main
CMD ["./vaultd"]
