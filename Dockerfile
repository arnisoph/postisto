FROM scratch
ADD build/postisto /posisto
CMD ["/posisto","--config","/config"]
