FROM alpine

RUN apk update && \
    apk add --no-cache tzdata

WORKDIR /app

ARG TARGETOS
ARG TARGETARCH

COPY ./bld/solaredge-$TARGETOS-$TARGETARCH /app/solaredge
ENTRYPOINT ["/app/solaredge"]
CMD []
