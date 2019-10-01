FROM golang:1.13-alpine
LABEL maintainer="jacobpatterson1549"
WORKDIR /go/src/github.com/jacobpatterson1549/nate-mlb
COPY . .
RUN go install
CMD ["nate-mlb"]