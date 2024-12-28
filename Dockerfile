# Build image
FROM golang:1.23.4 AS build
WORKDIR /project

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o api-server

# Distribution image
FROM ubuntu AS api-server
WORKDIR /project

COPY --from=build /project/api-server ./
COPY public ./public

EXPOSE 80
ENTRYPOINT [ "./api-server" ]