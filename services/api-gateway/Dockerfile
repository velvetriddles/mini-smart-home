FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY . .
RUN go mod download && CGO_ENABLED=0 go build -o /out/gateway ./cmd/gateway

FROM gcr.io/distroless/static-debian12
COPY --from=build /out/gateway /gateway
ENTRYPOINT ["/gateway"] 