FROM ubuntu

COPY bigfile /bigfile/

ENTRYPOINT ["/bigfile/bigfile"]