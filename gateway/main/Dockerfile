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
ADD /gateway /go/src/TaskManager/gateway
ADD /svcdiscovery /go/src/TaskManager/svcdiscovery
ADD /config /go/src/TaskManager/gateway/main/config

#Set workdir to import all packages
WORKDIR /go/src/TaskManager/gateway/main
RUN go get ./...

#Build
RUN go build -o gwd main.go
EXPOSE 8000 8000
CMD ["./gwd"]
