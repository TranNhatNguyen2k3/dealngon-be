FROM golang:1.26-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /out/api ./cmd/api && CGO_ENABLED=0 go build -o /out/worker ./cmd/worker

FROM alpine:3.22
RUN apk add --no-cache ca-certificates wget
WORKDIR /app
COPY --from=build /out/api /api
COPY --from=build /out/worker /worker
EXPOSE 8080
CMD ["/api"]
