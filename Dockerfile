# Self-contained build: the local dependency (ligneous-gedcom-lib) is vendored
# into vendor/, so the build context is just this repo — no sibling checkout needed.
FROM docker.io/library/golang:1.23-alpine AS builder
WORKDIR /src
COPY . .
RUN CGO_ENABLED=0 go build -mod=vendor -o /api .

FROM docker.io/library/alpine:3.19
RUN apk add --no-cache ca-certificates
EXPOSE 8091
ENV PORT=8091
COPY --from=builder /api /api
CMD ["/api", "-port", "8091"]
