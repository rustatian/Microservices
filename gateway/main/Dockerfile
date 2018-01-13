FROM golang
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD /gateway gateway
ADD /config gateway/main/config
EXPOSE 8000 8000
WORKDIR gateway/main
CMD ["./gwd"]
