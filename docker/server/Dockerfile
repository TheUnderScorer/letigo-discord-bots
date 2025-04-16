# --- Build Stage ---
FROM golang:1.24 AS build
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
COPY tools ./tools

# Download Go modules
RUN go mod download

# Build the production binaries
RUN cd src && bash ../scripts/build-prod.sh && cd ..

# --- Runtime Stage ---
FROM debian:bullseye-slim AS runtime
WORKDIR /app

# Install runtime dependencies
RUN apt-get -y update \
    && apt-get -y install -y ffmpeg ca-certificates \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Copy only the built output from the build stage
COPY --from=build /app/src ./src

# Expose the application's port
EXPOSE 3000

# Command to run the application
CMD ["./src/app"]