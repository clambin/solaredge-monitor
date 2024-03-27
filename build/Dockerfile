FROM alpine

RUN apk update && \
    apk add --no-cache tzdata

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH
ARG COMMAND

RUN echo $COMMAND "$COMMAND"
RUN chmod +x ./build/entrypoint-$COMMAND.sh ./bld/solaredge-$TARGETOS-$TARGETARCH

COPY ./build/entrypoint-$COMMAND.sh /app/entrypoint.sh
COPY ./bld/solaredge-$TARGETOS-$TARGETARCH /app/solaredge

EXPOSE 80
ENTRYPOINT ["/app/entrypoint.sh"]
CMD []
