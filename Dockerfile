FROM golang:1.27-alpine AS build
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
RUN go test ./... && CGO_ENABLED=0 go build -trimpath -o /out/zentraspace ./cmd/zentraspace
FROM alpine:3.22
WORKDIR /app
COPY --from=build /out/zentraspace /usr/local/bin/zentraspace
USER 65532:65532
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/zentraspace"]
