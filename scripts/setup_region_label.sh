#!/bin/sh
CLUSTER_REGION_LABEL=${CLUSTER_REGION_LABEL:-"topology.kubernetes.io/region=local"}

for n in $(kubectl get nodes -o jsonpath='{.items[].metadata.name}'); do
	kubectl get node $n -o jsonpath='{.metadata.labels}' 2> /dev/null \
		| grep -q "${CLUSTER_REGION_LABEL%=*}"
	if test $? -ne 0; then
		kubectl label node $n "$CLUSTER_REGION_LABEL";
	fi
done
