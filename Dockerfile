# builder image
FROM golang:1.13-alpine3.11 as builder

RUN mkdir /build
ADD cmd/web/*.go /build/
ADD go.mod /build/
WORKDIR /build
RUN CGO_ENABLED=0 GOOS=linux go build -a -o app .

# generate clean, final image for end users
FROM alpine:3.11.3
COPY --from=builder /build/app .
EXPOSE 4000

CMD ["./app"]