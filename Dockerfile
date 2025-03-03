FROM golang AS builder

WORKDIR /app

COPY . .

RUN go get .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app .

# deployment image
FROM alpine:latest  
RUN apk --no-cache add ca-certificates

WORKDIR /app
COPY --from=builder /app/app .

CMD [ "./app" ]