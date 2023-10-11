FROM golang:1.20-alpine AS build-stage

WORKDIR /app
COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /goapp

FROM gcr.io/distroless/static-debian12 AS build-release-stage

WORKDIR /

COPY --from=build-stage /goapp /goapp

EXPOSE 8000

USER nonroot:nonroot

CMD ["/goapp"]
