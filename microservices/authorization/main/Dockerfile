FROM golang
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD /microservices/authorization authorization
ADD /config authorization/main/config
EXPOSE 10001 10001
EXPOSE 3306
EXPOSE 6379
WORKDIR authorization/main
CMD ["./authd"]
