#########################################################################################################
# Frontend Build
#########################################################################################################
# Container qualification builds on native GitHub amd64 and arm64 runners.

FROM node:26.13.0 AS frontend-build
ARG FASTEN_ENV=sandbox
WORKDIR /usr/src/fastenhealth/frontend
COPY frontend/package.json frontend/yarn.lock ./

RUN npm exec --yes --package=corepack@0.35.0 -- corepack yarn install --frozen-lockfile --registry https://registry.npmjs.org --network-concurrency 4
COPY frontend/ ./
RUN --mount=type=cache,target=/tmp/lock,sharing=locked \
    node node_modules/@angular/cli/bin/ng.js build --configuration "${FASTEN_ENV}" --optimization=true --build-optimizer=true --output-path=../dist

#########################################################################################################
# Backend Build
#########################################################################################################
FROM golang:1.27.3-bookworm AS backend-build

WORKDIR /go/src/github.com/fastenhealth/fasten-onprem
COPY . .

RUN --mount=type=cache,target=/tmp/lock,sharing=locked \
    --mount=type=cache,target=/go/pkg/mod,sharing=locked \
    --mount=type=cache,target=/root/.cache/go-build,sharing=locked \
    go mod vendor \
    && go install github.com/golang/mock/mockgen@v1.6.0 \
    && go generate ./... \
    && go vet ./... \
    && go test -timeout=20m ./... \
    && CGO_ENABLED=1 go build -ldflags "-extldflags=-static" -tags "static" -o /go/bin/fasten ./backend/cmd/fasten/ \
    && CGO_ENABLED=1 go build -ldflags "-extldflags=-static" -tags "static" -o /go/bin/fasten-backup ./backend/cmd/fasten-backup/

# create folder structure
RUN mkdir -p /opt/fasten/db \
  && mkdir -p /opt/fasten/web \
  && mkdir -p /opt/fasten/config

#########################################################################################################
# Distribution Build
#########################################################################################################
FROM gcr.io/distroless/static-debian11

EXPOSE 8080
WORKDIR /opt/fasten/
COPY --from=backend-build  /opt/fasten/ /opt/fasten/
COPY --from=frontend-build /usr/src/fastenhealth/dist /opt/fasten/web
COPY --from=backend-build /go/bin/fasten /opt/fasten/fasten
COPY --from=backend-build /go/bin/fasten-backup /opt/fasten/fasten-backup
COPY LICENSE.md /opt/fasten/LICENSE.md
COPY config.yaml /opt/fasten/config/config.yaml
RUN ["/opt/fasten/fasten", "--help"]
RUN ["/opt/fasten/fasten-backup", "--help"]
CMD ["/opt/fasten/fasten", "start", "--config", "/opt/fasten/config/config.yaml"]
