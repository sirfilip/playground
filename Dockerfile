FROM golang:1.24.2-alpine3.21 as builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download


COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /proj

FROM alpine:3.21

WORKDIR /

COPY --from=builder /proj /proj

EXPOSE 5000 

# USER nonroot:nonroot

CMD ["/proj"]
