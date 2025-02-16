FROM golang:1.23 AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -tags netgo -ldflags '-s -w' -o trip .

FROM scratch
COPY --from=builder /app/trip .
COPY --from=builder /app/swagger.yaml .
EXPOSE $API_PORT
CMD ["./trip"]
