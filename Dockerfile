FROM golang:1.16 as build
WORKDIR /usr/auth-server/
COPY go.* ./
RUN go mod download

RUN go run github.com/prisma/prisma-client-go prefetch
COPY schema.prisma ./
RUN mkdir db && go run github.com/prisma/prisma-client-go generate

COPY . .
RUN go build -o ./auth-server

FROM golang:1.16 as release
WORKDIR /root/
COPY --from=build /usr/auth-server/auth-server ./
COPY rsa.private ./
ENTRYPOINT [ "./auth-server" ]