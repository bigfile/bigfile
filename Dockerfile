FROM ubuntu

COPY bigfile /bigfile/

WORKDIR /bigfile

ENTRYPOINT ["/bigfile/bigfile"]