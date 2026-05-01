FROM golang:1.24 AS builder
WORKDIR /workspace
COPY operator/go.mod operator/go.sum ./
RUN go mod download
COPY operator/ ./
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o manager .

FROM gcr.io/distroless/static:nonroot
WORKDIR /
COPY --from=builder /workspace/manager /manager
USER 65532:65532
ENTRYPOINT ["/manager"]
