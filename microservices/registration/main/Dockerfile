FROM golang
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD /microservices/registration registration
ADD /config registration/main/config
EXPOSE 10002 10002
EXPOSE 3306
WORKDIR calendar/main
CMD ["./regd"]