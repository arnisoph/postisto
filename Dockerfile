FROM alpine
ARG TARGETOS
ARG TARGETARCH
RUN \
  apk add --no-cache \
    ca-certificates
ADD build/postisto-*.${TARGETOS}-${TARGETARCH}.tar.gz /
ENTRYPOINT ["/postisto"]
CMD ["--config","/config"]
