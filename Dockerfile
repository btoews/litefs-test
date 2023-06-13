FROM flybento/litefs:latest AS litefs

# fetch deps
FROM golang:alpine AS builder
RUN apk add --no-cache --update gcc libc-dev
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
RUN go build github.com/mattn/go-sqlite3

# build
COPY . ./
RUN go build -o foo .

FROM alpine:latest AS runner
WORKDIR /
RUN apk add --no-cache --update fuse3
COPY --from=litefs /usr/local/bin/litefs /usr/local/bin/litefs
COPY --from=builder /build/foo /usr/local/bin/foo
COPY ./etc/ /etc/
CMD ["litefs", "mount"]
