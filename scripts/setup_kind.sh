#!/bin/sh
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

if test -z "$(docker inspect -f='{{json .NetworkSettings.Networks.kind}}' $LOCAL_REG_NAME)"; then
	docker network connect "kind" $LOCAL_REG_NAME
fi
