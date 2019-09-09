FROM scratch

COPY bigfile /

ENTRYPOINT ["/bigfile"]