FROM golang:1.23
WORKDIR /app

RUN apt-get -y update
RUN apt-get -y upgrade
RUN apt-get install -y ffmpeg

COPY go.work go.work.sum ./
COPY src-go ./src-go
COPY tools ./tools
RUN go mod download

RUN cd src-go && CGO_ENABLED=0 GOOS=linux go build -o /src-go

EXPOSE 8080
