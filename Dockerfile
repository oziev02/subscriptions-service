# ---- Build stage ----
FROM golang:1.23 as builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /bin/app ./cmd/subscriptions

# ---- Runtime stage ----
FROM gcr.io/distroless/base-debian12
WORKDIR /
COPY --from=builder /bin/app /bin/app
COPY ./.env ./.env
COPY ./docs/openapi.yaml /docs/openapi.yaml
ENTRYPOINT ["/bin/app"]
