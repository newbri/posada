FROM golang:latest AS builder
WORKDIR /app
COPY . .
COPY app.yaml .
RUN go get -d -v ./... # Download and install the dependencies
RUN go build -o main main.go
COPY db/migration ./db/migration

EXPOSE 8080
CMD [ "/app/main" ]