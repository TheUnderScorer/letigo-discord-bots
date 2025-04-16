# --- Build Stage ---
FROM letigo-discord-bots-base:latest AS build
WORKDIR /app

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