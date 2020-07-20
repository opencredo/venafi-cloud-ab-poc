FROM golang:buster AS build
ARG GIT_COMMIT_ID=""

RUN mkdir -p /src
COPY . /src
WORKDIR /src

RUN make bin/secretfixer

FROM gcr.io/distroless/base
COPY --from=build /src/bin/secretfixer .
ENV OCVAB_SECRETFIXER_LISTEN=":8080"
EXPOSE 8080/tcp

CMD [ "./secretfixer" ]
