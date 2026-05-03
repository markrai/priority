# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/priority .

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=build /out/priority /app/priority
COPY public/ /app/public/
RUN mkdir -p /app/data
ENV DATA_FILE=/app/data/projects.json \
	PORT=8080
EXPOSE 8080
CMD ["/app/priority"]
