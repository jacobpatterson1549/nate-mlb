FROM golang:1.13-alpine AS build

WORKDIR /app

# fetch dependencies first so they will not have to be refetched when other source code changes
COPY go.mod go.sum /app/

RUN go mod download

COPY . /app/

# build web assembly
RUN GOOS=js GOARCH=wasm go build -o /app/static/main.wasm go/cmd/wasm/main.go

# build server without links to C libraries
RUN CGO_ENABLED=0 go build -o /app/nate-mlb go/cmd/server/main.go

FROM scratch

# copy the x509 certificate file for Alpine Linux to allow server to make https requests
COPY --from=build /etc/ssl/cert.pem /etc/ssl/cert.pem

COPY --from=build /app /

COPY --from=build "/usr/local/go/misc/wasm/wasm_exec.js" /js/

# use exec form to not run from shell, which scratch image does not have
CMD ["/nate-mlb"]