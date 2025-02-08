FROM golang:1.23.4-alpine3.21 as build
WORKDIR /workspace
COPY go.mod go.sum ./
RUN go mod download -x
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o server .

FROM alpine:3.21.2 as run
WORKDIR /workspace
COPY --from=build /workspace/server .
CMD ./server
EXPOSE 1323