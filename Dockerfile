FROM alpine:3.14

COPY sqpr /
EXPOSE 8080

ENTRYPOINT [ "/sqpr" ]
CMD [ "server", "run", "--port", "8080" ]
