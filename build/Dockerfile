FROM alpine

RUN apk update && \
    apk add --no-cache tzdata

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

COPY ./bld/solaredge-monitor-$TARGETOS-$TARGETARCH /app/solaredge-monitor

EXPOSE 80
ENTRYPOINT ["/app/solaredge-monitor"]
CMD []
