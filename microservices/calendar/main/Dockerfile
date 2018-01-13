FROM golang
MAINTAINER Valery Piashchynski, Adexin
ENV DEV False
ADD /microservices/calendar calendar
ADD /config calendar/main/config
EXPOSE 10003 10003
WORKDIR calendar/main
CMD ["./tcald"]