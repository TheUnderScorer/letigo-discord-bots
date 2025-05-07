FROM golang:1.24
WORKDIR /app

# Update and install necessary dependencies
RUN apt-get -y update \
    && apt-get -y upgrade \
    && apt-get install -y jq

# Copy necessary files for the build
COPY package.json ./
COPY scripts ./scripts
COPY go.work go.work.sum ./
COPY src ./src
COPY lib ./lib
COPY wojciech-bot ./wojciech-bot
COPY tools ./tools

# Download Go modules
RUN go mod download