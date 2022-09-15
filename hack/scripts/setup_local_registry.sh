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

case "$(docker inspect -f '{{.State.Running}}' $LOCAL_REG_NAME 2> /dev/null)" in
	"false")
			echo "Starting registry <$LOCAL_REG_NAME>"
			docker start "$LOCAL_REG_NAME"
			;;
		"")
			echo "Setting up registry <$LOCAL_REG_NAME>"
			docker run \
				--detach=true \
				--restart=always \
				--publish "127.0.0.1:$REG_PORT:5000" \
				--name "$LOCAL_REG_NAME" \
				registry:2 > /dev/null
			;;
esac

echo -e "Registry <$LOCAL_REG_NAME> is ready\n"
