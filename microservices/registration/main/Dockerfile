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
ADD /microservices/registration /go/src/TaskManager/microservices/registration
ADD /svcdiscovery /go/src/TaskManager/svcdiscovery
ADD /config /go/src/TaskManager/microservices/registration/main/config

#Set workdir to import all packages
WORKDIR /go/src/TaskManager/microservices/registration/main
RUN go get ./...

#Build
RUN go build -o regd main.go

EXPOSE 10002 10002
EXPOSE 3306

CMD ["./regd"]