# Build image
FROM golang:1.23.4 AS build
WORKDIR /project

COPY go.mod go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o timeseries-api

# Distribution image
FROM ubuntu AS timeseries-api
WORKDIR /project

COPY --from=build /project/timeseries-api ./
COPY public ./public

EXPOSE 80
ENTRYPOINT [ "./timeseries-api" ]