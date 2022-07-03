FROM node:16.15.0 as builder
WORKDIR /app

COPY . ./

RUN npm ci
RUN npm run server:build

FROM node:16.15.0

WORKDIR /app

COPY --from=builder /app/dist/server/src .
COPY --from=builder /app/package.json .
COPY --from=builder /app/package-lock.json .

RUN npm ci

CMD ["node", "server/server.js"]
