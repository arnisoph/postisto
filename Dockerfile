FROM alpine
RUN \
  apk add --no-cache \
    ca-certificates
COPY build/postisto /postisto
ENTRYPOINT ["/postisto"]
CMD ["--config","/config"]
