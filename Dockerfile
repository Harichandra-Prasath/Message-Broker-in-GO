FROM golang:1.21

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY *.go /app/

COPY Makefile /app/

EXPOSE 3000 4000 5000
CMD [ "make" ]