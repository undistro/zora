#!/bin/sh
# Copyright 2022 Undistro Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

set -o errexit

LOCAL_REG_NAME=${LOCAL_REG_NAME:-"local_registry"}
REG_PORT=${REG_PORT:-5000}
REG_HOST=${REG_HOST:-"localhost"}
REG_ADDR=${REG_ADDR:-"$REG_HOST:$REG_PORT"}

cat <<EOF | kind create cluster --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."$REG_ADDR"]
    endpoint = ["http://$LOCAL_REG_NAME:$REG_PORT"]
EOF

if ! docker inspect -f='{{json .NetworkSettings.Networks.kind}}' $LOCAL_REG_NAME > /dev/null; then
	docker network connect "kind" $LOCAL_REG_NAME
fi
