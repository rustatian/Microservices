FROM golang
MAINTAINER Valery Piashchynski, Adexin
#Dev or production
ENV DEV False

#Buil for linux
ENV GOOS linux
ENV GOARCH amd64

#Create dir in GOPATH
RUN mkdir /go/src/TaskManager

#Add Dep
ADD /microservices/calendar /go/src/TaskManager/microservices/calendar
ADD /svcdiscovery /go/src/TaskManager/svcdiscovery
ADD /config /go/src/TaskManager/microservices/calendar/main/config

#Set workdir to import all packages
WORKDIR /go/src/TaskManager/microservices/calendar/main
RUN go get ./...

#Build
RUN go build -o tcald main.go

EXPOSE 10003 10003

CMD ["./tcald"]