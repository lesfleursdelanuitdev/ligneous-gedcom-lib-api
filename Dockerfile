# Build from parent so we can access ligneous-gedcom-lib
FROM golang:1.23-alpine AS builder
WORKDIR /build

# Copy lib first (dependency)
COPY ligneous-gedcom-lib/ ./ligneous-gedcom-lib/
COPY ligneous-gedcom-lib-api/ ./ligneous-gedcom-lib-api/

# Build (replace directive in go.mod expects ../ligneous-gedcom-lib)
WORKDIR /build/ligneous-gedcom-lib-api
RUN go mod download && CGO_ENABLED=0 go build -o /api .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
EXPOSE 8092
ENV PORT=8092
COPY --from=builder /api /api
CMD ["/api", "-port", "8092"]
