# syntax=docker/dockerfile:1
FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /slam-team-api ./cmd/api
RUN CGO_ENABLED=0 go build -o /slamctl ./cmd/slamctl

FROM alpine:3.20
RUN apk add --no-cache ca-certificates tzdata
COPY --from=build /slam-team-api /slam-team-api
COPY --from=build /slamctl /slamctl
EXPOSE 8080
ENTRYPOINT ["/slam-team-api"]
