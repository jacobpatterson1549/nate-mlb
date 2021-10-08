FROM golang:1.17-alpine3.14 AS build

WORKDIR /app

# fetch dependencies first so they will not have to be refetched when other source code changes
COPY go.mod go.sum /app/

RUN go mod download

COPY . /app/

# build application without links to C libraries
RUN CGO_ENABLED=0 go test ./... --cover \
    && CGO_ENABLED=0 go build -o /app/nate-mlb

FROM scratch

# copy the x509 certificate file for Alpine Linux to allow server to make https requests
COPY --from=build /etc/ssl/cert.pem /etc/ssl/cert.pem

COPY --from=build /app/nate-mlb /

# use exec form to not run from shell, which scratch image does not have
CMD ["/nate-mlb"]