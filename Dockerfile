FROM golang:1.13-alpine AS builder

WORKDIR /app

# fetch dependencies first so they will not have to be refetched when other source code changes
COPY go.mod go.sum ./

RUN go mod download

COPY . .

# build application without links to C libraries
RUN CGO_ENABLED=0 go build -o /app/nate-mlb

FROM scratch

# copy the x509 certificate file for Alpine Linux
COPY --from=builder /etc/ssl/cert.pem /etc/ssl/cert.pem

WORKDIR /app

COPY --from=builder /app /app/

# use exec form to not run from shell, which scratch image does not have
ENTRYPOINT ["/app/nate-mlb"]