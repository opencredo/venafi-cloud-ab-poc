FROM golang:buster AS build
ARG GIT_COMMIT_ID=""

RUN mkdir -p /src
COPY . /src
WORKDIR /src

RUN make bin/txnsim

FROM gcr.io/distroless/base
COPY --from=build /src/bin/txnsim .
ENV OCVAB_TXNSIM_LISTEN=":8080"
CMD [ "./txnsim" ]
