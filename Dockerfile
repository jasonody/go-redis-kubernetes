# Dockerfile References: https://docs.docker.com/engine/reference/builder/

### Build stage ###

FROM golang:latest as builder

LABEL maintainer="Jason Ody <jasonody@users.noreply.github.com>"

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

# -a: force rebuilding of packages that are already up-to-date
# -installsuffix: a suffix to use in the name of the package installation directory, in order to keep output separate from default builds
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


### App stage ###

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

EXPOSE 8080

CMD ["./main"]