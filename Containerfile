FROM golang:latest AS build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o enclave-wizard .

FROM alpine:latest
WORKDIR /app
COPY --from=build /app/enclave-wizard .
EXPOSE 8080
ENTRYPOINT ["./enclave-wizard"]
