FROM golang
MAINTAINER Valery Piashchynski
#Dev or production
ENV DEV False

#Buil for linux
ENV GOOS linux
ENV GOARCH amd64

#Create dir in GOPATH
RUN mkdir /go/src/TaskManager

#Add Dep
ADD /microservices/authorization /go/src/TaskManager/microservices/authorization
ADD /svcdiscovery /go/src/TaskManager/svcdiscovery
ADD /config /go/src/TaskManager/microservices/authorization/main/config

#Set workdir to import all packages
WORKDIR /go/src/TaskManager/microservices/authorization/main
RUN go get ./...

#Build
RUN go build -o authd main.go

EXPOSE 10001 10001
#?? Delete
EXPOSE 5432
EXPOSE 6379

CMD ["./authd"]
CMD ["./1d"]
CMD ["./2d"]
CMD ["./3d"]
CMD ["./4d"]


