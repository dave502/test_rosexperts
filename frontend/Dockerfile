FROM golang:1.14 as builder
COPY . .
RUN go build \
  -ldflags "-linkmode external -extldflags -static" \
  -o server .

### run stage ###
FROM scratch
COPY --from=builder go/server ./server
EXPOSE 8888
CMD ["./server"]"
