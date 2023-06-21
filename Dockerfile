FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
RUN go mod download
COPY go.sum ./

COPY *.go ./

RUN go build .

EXPOSE 8080

CMD [ "./main" ]