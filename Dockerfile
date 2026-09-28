# syntax=docker/dockerfile:1

FROM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci
COPY web/ ./
RUN npm run build -- --outDir /src/internal/web/dist --emptyOutDir

FROM golang:1.27-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=web /src/internal/web/dist ./internal/web/dist
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/big-brother ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/big-brother /big-brother
EXPOSE 8080
VOLUME ["/data"]
ENV BB_DB_PATH=/data/big-brother.db
ENTRYPOINT ["/big-brother"]
