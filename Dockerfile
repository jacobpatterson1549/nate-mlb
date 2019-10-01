FROM golang:1.13-alpine AS builder

WORKDIR /app

# fetch dependencies first so they will not have to be refetched when other source code changes
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/nate-mlb

FROM alpine:3.10

WORKDIR /app

COPY --from=builder /app /app/

CMD /app/nate-mlb