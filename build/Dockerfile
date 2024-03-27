FROM alpine

RUN apk update && \
    apk add --no-cache tzdata

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH
#ARG COMMAND

#COPY ./build/entrypoint-$COMMAND.sh /app/entrypoint.sh
#COPY ./bld/solaredge-$TARGETOS-$TARGETARCH /app/solaredge
#RUN chmod +x /app/entrypoint.sh
#RUN chmod +x /app/solaredge
#ENTRYPOINT ["/app/entrypoint.sh"]
#CMD []

COPY ./bld/solaredge-$TARGETOS-$TARGETARCH /app/solaredge
ENTRYPOINT ["/app/solaredge"]
CMD []
