# --- Build Stage ---
FROM letigo-discord-bots-base:latest AS build
WORKDIR /app

# Build the production binaries
RUN cd wojciech-bot && make build_prod_linux

# --- Runtime Stage ---
FROM debian:bullseye-slim AS runtime
WORKDIR /app

# Install runtime dependencies
RUN apt-get -y update \
    && apt-get -y install -y ca-certificates ffmpeg curl python3 \
    && apt-get clean \
    && curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp \
    && rm -rf /var/lib/apt/lists/*

# Copy only the built output from the build stage
COPY --from=build /app/wojciech-bot/wojciech-bot ./

# Expose the application's port
EXPOSE 3000

# Command to run the application
CMD ["./wojciech-bot"]