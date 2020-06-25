FROM golang:buster AS build
ARG GIT_COMMIT_ID=""

RUN mkdir -p /src
COPY . /src
WORKDIR /src

RUN make bin/ledger

FROM gcr.io/distroless/base
COPY --from=build /src/bin/ledger .
ENV OCVAB_LEDGER_LISTEN=":8080"
CMD [ "./ledger" ]
