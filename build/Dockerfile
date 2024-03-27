FROM alpine

RUN apk update && \
    apk add --no-cache tzdata

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH
ARG COMMAND

RUN echo $COMMAND "$COMMAND"

COPY ./build/entrypoint-$COMMAND.sh /app/entrypoint.sh
COPY ./bld/solaredge-$TARGETOS-$TARGETARCH /app/solaredge
RUN chmod +x /app/*

ENTRYPOINT ["/app/entrypoint.sh"]
CMD []
