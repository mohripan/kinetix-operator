FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/producer ./cmd/producer && \
    CGO_ENABLED=0 go build -o /out/worker ./cmd/worker && \
    CGO_ENABLED=0 go build -o /out/sink ./cmd/sink

FROM alpine:3.21
RUN adduser -D -u 10001 kinetix
USER kinetix
COPY --from=build /out/producer /producer
COPY --from=build /out/worker /worker
COPY --from=build /out/sink /sink
CMD ["/worker"]
