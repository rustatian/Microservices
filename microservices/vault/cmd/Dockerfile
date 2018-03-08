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
ADD /microservices/vault /go/src/TaskManager/microservices/vault
ADD /svcdiscovery /go/src/TaskManager/svcdiscovery
ADD /config /go/src/TaskManager/microservices/vault/main/config

#Set workdir to import all packages
WORKDIR /go/src/TaskManager/microservices/vault/main
RUN go get ./...

#Build
RUN go build -o vaultd main.go

EXPOSE 10000 10000

CMD ["./vaultd"]