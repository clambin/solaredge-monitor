FROM alpine

RUN apk update && \
    apk add --no-cache tzdata

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH
ARG COMMAND

RUN find /app -type f -print
COPY ./build/entrypoint-$COMMAND.sh /app/entrypoint.sh
COPY ./bld/binaries/solaredge-$TARGETOS-$TARGETARCH /app/solaredge
RUN chmod +x /app/entrypoint.sh
RUN chmod +x /app/solaredge
RUN find /app -type f -print

ENTRYPOINT ["/app/entrypoint.sh"]
CMD []
