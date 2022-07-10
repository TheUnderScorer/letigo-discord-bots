ARG IMAGE=node:16-alpine3.15

FROM --platform=linux/amd64 $IMAGE as builder
WORKDIR /app

COPY . ./

RUN npm ci
RUN npm run server:build

FROM --platform=linux/amd64 $IMAGE

WORKDIR /app

COPY --from=builder /app/dist/server/src .
COPY --from=builder /app/package.json .
COPY --from=builder /app/package-lock.json .

RUN apk add --update ffmpeg
RUN npm ci

CMD ["node", "server/server.js"]
