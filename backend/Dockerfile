FROM golang:1.26-alpine AS builder

WORKDIR /workspace/backend
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go tool revel build -a=. -t=./dist -m=prod

FROM alpine:3.22 AS runner

RUN apk add --no-cache ca-certificates \
    && addgroup -S app \
    && adduser -S -G app app
WORKDIR /app
COPY --from=builder --chown=app:app /workspace/backend/dist/ ./
RUN chmod +x /app/run.sh /app/backend
USER app
EXPOSE 9000
CMD ["./run.sh"]
