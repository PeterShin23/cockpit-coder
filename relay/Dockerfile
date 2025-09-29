FROM golang:1.22-alpine AS build
WORKDIR /src
COPY . .
RUN go build -o /out/relay ./cmd/relay

FROM alpine:3.20
WORKDIR /app
COPY --from=build /out/relay /app/relay
ENV PORT=8081
EXPOSE 8081
CMD ["/app/relay"]
