FROM index.docker.io/library/golang:1.23.4-alpine3.21 AS golang
FROM index.docker.io/library/alpine:3.21.2 AS alpine

FROM golang as build
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server .

FROM alpine as run
WORKDIR /workspace
COPY --from=build /workspace/server .
CMD ./server
EXPOSE 1323
