FROM ubuntu

RUN apt-get install libgraphicsmagick1-dev

COPY bigfile /bigfile/

WORKDIR /bigfile

ENTRYPOINT ["/bigfile/bigfile"]