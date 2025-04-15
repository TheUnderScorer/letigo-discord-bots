FROM golang:1.24
WORKDIR /app

RUN apt-get -y update
RUN apt-get -y upgrade
RUN apt-get install -y ffmpeg jq

COPY package.json ./
COPY scripts ./scripts
COPY go.work go.work.sum ./
COPY src ./src
COPY tools ./tools
COPY deploy ./deploy

RUN go mod download

RUN cd src && bash ../scripts/build-prod.sh && cd ..

EXPOSE 3000

CMD ["./src/app"]
