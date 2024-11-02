FROM golang:1.23

WORKDIR /app

RUN apt-get -y update
RUN apt-get -y upgrade
RUN apt-get install -y ffmpeg

COPY ./tools /app/tools
COPY src /app/src
COPY ./go.work ./go.work.sum /app/

RUN go mod download
