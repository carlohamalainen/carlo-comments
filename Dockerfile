FROM --platform=linux/amd64 golang:alpine3.20 AS build
WORKDIR /app
COPY api/ .
RUN go build -o main .

FROM alpine
WORKDIR /app
COPY --from=build /app/main .
CMD ["./main"]