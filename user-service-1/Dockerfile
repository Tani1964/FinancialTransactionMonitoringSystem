FROM golang:1.24-alpine
WORKDIR /app
COPY . .
RUN go mod tidy && go build -o fms-user-service-1 main.go
EXPOSE 3000
CMD ["./fms-user-service-1"]