# Copyright 2025 Undistro Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

FROM --platform=$BUILDPLATFORM goreleaser/goreleaser:v2.17.1 AS builder
ARG TARGETOS=linux
ARG TARGETARCH
ARG REPO="https://github.com/aquasecurity/trivy.git"
ARG BRANCH=v0.74.0
ARG EXPECTED_COMMIT=e1fd17a0ea4a8cf24bc4b4dd7e2cfbf4bb31b994

WORKDIR /
RUN git clone --recurse-submodules --shallow-submodules --depth 1 ${REPO} --branch ${BRANCH}
WORKDIR /trivy
RUN test "$(git rev-parse HEAD)" = "$EXPECTED_COMMIT"

COPY patches patches
RUN git apply patches/*.patch

RUN GOOS="$TARGETOS" GOARCH="$TARGETARCH" goreleaser build --id build-linux --single-target --clean --skip=validate -o dist/trivy

FROM alpine:3.24

RUN apk update && apk upgrade && \
    apk --no-cache add ca-certificates git && \
    rm -rf /var/cache/apk

COPY --from=builder /trivy/dist/trivy /usr/local/bin/trivy.bin
COPY --from=builder /trivy/contrib/*.tpl contrib/

RUN echo '#!/bin/sh' > /usr/local/bin/trivy && \
    echo 'umask 002 && echo "umask set to $(umask)" && /usr/local/bin/trivy.bin "$@"' >> /usr/local/bin/trivy && \
    chmod +x /usr/local/bin/trivy

RUN adduser -u 5000 -D nonroot && addgroup nonroot root
USER 5000
ENTRYPOINT ["trivy"]
