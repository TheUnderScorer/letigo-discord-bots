FROM --platform=linux/amd64 ghcr.io/coqui-ai/tts-cpu

WORKDIR /app

COPY requirements.txt /app
RUN --mount=type=cache,target=/root/.cache/pip \
    pip3 install -r requirements.txt

COPY . /app

ENV COQUI_TOS_AGREED=1

# Download model first
#RUN python3 main.py --init

ENTRYPOINT ["python3"]
CMD ["main.py"]
