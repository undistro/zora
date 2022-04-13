#!/bin/sh
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
				--publish "$REG_PORT:5000" \
				--name "$LOCAL_REG_NAME" \
				registry:2 &> /dev/null
			;;
esac

echo "Registry <$LOCAL_REG_NAME> is ready\n"
