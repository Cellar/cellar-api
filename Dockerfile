FROM  golang:1.23-alpine AS build

WORKDIR /cellar

COPY go.mod go.sum ./

RUN go mod download

COPY . .

ARG APP_VERSION

RUN go build -o /out/cellar -ldflags="-X main.version=${APP_VERSION}" cellar/cmd/cellar

FROM alpine:3

ARG GID=9001
ARG GROUP=cellar
ARG UID=9001
ARG USER=cellar

RUN addgroup -S -g $GID $GROUP && \
    adduser -DS -h /app -u $UID -G $GROUP $USER

USER $USER

WORKDIR /app

COPY --from=build \
     --chown=$USER \
     /out/cellar .

COPY --chown=$USER \
     docker-entrypoint.sh ./docker-entrypoint

RUN chmod +x docker-entrypoint

ENV DISABLE_SWAGGER=true
ENV APP_BIND_ADDRESS=:8080
ENV GIN_MODE=release


EXPOSE 8080

ENTRYPOINT [ "./docker-entrypoint" ]
