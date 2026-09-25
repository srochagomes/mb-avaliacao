FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 go build -o /out/server ./cmd/server

# alpine:3.17 já costuma existir em caches locais; 3.20 também serve
FROM alpine:3.17
RUN apk add --no-cache ca-certificates wget
COPY --from=build /out/server /server
EXPOSE 8081
USER nobody
ENTRYPOINT ["/server"]
