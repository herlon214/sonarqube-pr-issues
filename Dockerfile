################################
# Build binary
################################
FROM golang:1.17 as build
WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /app/sqpr ./main.go

################################
# Execute
################################
FROM alpine:3.14
COPY --from=build /app/sqpr /app/sqpr

ENTRYPOINT [ "./app/sqpr" ]
EXPOSE 8080
CMD [ "server", "run", "--port", "8080" ]
