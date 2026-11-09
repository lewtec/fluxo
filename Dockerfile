# Runtime image for GoReleaser (dockers_v2).
# Binary is copied from the build context as $TARGETPLATFORM/fluxo.

FROM alpine:3.24@sha256:f59abe656b692d901e184be3445e084786602108599d2a51d229c0f5a48c71bf

RUN apk add --no-cache ca-certificates

ARG TARGETPLATFORM
COPY $TARGETPLATFORM/fluxo /usr/local/bin/fluxo

ENV FLUXO_CONTAINER=1 \
    FLUXO_API_HOST=0.0.0.0 \
    FLUXO_API_PORT=8080 \
    FLUXO_DATA_DIR=/data/downloads \
    FLUXO_DATABASE=/data/session.db

EXPOSE 8080

ENTRYPOINT ["/usr/local/bin/fluxo"]
