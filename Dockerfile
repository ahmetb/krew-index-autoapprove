FROM golang:1-alpine AS build
RUN apk add --no-cache git
COPY bumpctl .
RUN CGO_ENABLED=0 go build -o /server ./webhook

FROM alpine
RUN apk add --no-cache ca-certificates
COPY --from=build /server /server
ENTRYPOINT ["/server"]
