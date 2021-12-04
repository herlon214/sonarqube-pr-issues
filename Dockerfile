FROM alpine:3.14

ENTRYPOINT [ "/sqpr" ]
EXPOSE 8080
CMD [ "server", "run", "--port", "8080" ]
