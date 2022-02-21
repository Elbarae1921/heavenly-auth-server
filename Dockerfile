FROM golang:1.17 as build
WORKDIR /usr/auth-server/
COPY go.* ./
RUN go mod download

RUN go run github.com/prisma/prisma-client-go prefetch
COPY schema.prisma ./
RUN mkdir db && go run github.com/prisma/prisma-client-go generate

COPY . .
ENV CGO_ENABLED=0

RUN GOOS=linux go build -o ./auth-server

FROM alpine:latest as release
WORKDIR /app
COPY --from=build /usr/auth-server/auth-server ./
COPY rsa.private ./