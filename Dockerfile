FROM scratch
ADD build/postisto /postisto
ENTRYPOINT ["/postisto"]
CMD ["--config","/config"]
