# syntax=docker/dockerfile:1.7
FROM node:22.18.0-bookworm-slim AS frontend
WORKDIR /src
COPY package.json yarn.lock gulpfile.js webpack.config.js .babelrc ./
RUN corepack enable && yarn install --frozen-lockfile
COPY static ./static
RUN yarn gulp

FROM golang:1.25.12-bookworm AS backend
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/ethphish . \
    && CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/postgres-schema-prepare ./cmd/postgres-schema-prepare

FROM backend AS test
ARG TEST_PACKAGES=./...
RUN test -z "$(gofmt -l .)" \
    && go vet ./... \
    && go test ${TEST_PACKAGES}

FROM debian:bookworm-slim AS runtime
RUN apt-get update \
    && apt-get install --no-install-recommends -y ca-certificates python3 \
    && rm -rf /var/lib/apt/lists/* \
    && groupadd --system --gid 10001 ethphish \
    && useradd --system --uid 10001 --gid ethphish --home-dir /opt/ethphish ethphish

WORKDIR /opt/ethphish
ENV ETHPHISH_RUNTIME_ENV=production
COPY --from=backend /out/ethphish ./ethphish
COPY --from=backend /out/postgres-schema-prepare ./postgres-schema-prepare
COPY --from=backend /src/config.json /src/VERSION /src/ANGLERPHISH_VERSION /src/LICENSE ./
COPY --from=backend /src/db/db_postgres ./db/db_postgres
COPY --from=backend /src/templates ./templates
COPY --from=backend /src/reports/python/requirements.txt ./reports/python/requirements.txt
COPY --from=backend /src/static ./static
COPY --from=frontend /src/static/js/dist ./static/js/dist
COPY --from=frontend /src/static/css/dist ./static/css/dist
RUN mkdir -p /var/lib/ethphish/reports \
    && chown -R ethphish:ethphish /opt/ethphish /var/lib/ethphish

USER 10001:10001
EXPOSE 3333 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
  CMD python3 -c "import ssl, urllib.request; urllib.request.urlopen('https://127.0.0.1:3333/healthz', context=ssl._create_unverified_context(), timeout=2).read()"
ENTRYPOINT ["/opt/ethphish/ethphish"]
CMD ["--config", "/opt/ethphish/config.json"]
