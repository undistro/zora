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

CLUSTER_REGION_LABEL=${CLUSTER_REGION_LABEL:-"topology.kubernetes.io/region=local"}

for n in $(kubectl get nodes -o jsonpath='{.items[].metadata.name}'); do
	kubectl get node $n -o jsonpath='{.metadata.labels}' 2> /dev/null \
		| grep -q "${CLUSTER_REGION_LABEL%=*}"
	if test $? -ne 0; then
		kubectl label node $n "$CLUSTER_REGION_LABEL";
	fi
done
