FROM letigo-discord-bots-base:latest
WORKDIR /app

# Install runtime dependencies
RUN apt-get -y update \
    && apt-get -y install -y ca-certificates ffmpeg curl python3 \
    && apt-get clean \
    && curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp \
    && chmod a+rx /usr/local/bin/yt-dlp \
    && rm -rf /var/lib/apt/lists/* 