#!/usr/bin/env bash

GRAFANA_TAG=${1:-}

if echo "$GRAFANA_TAG" | grep -q "^v"; then
	_grafana_version=$(echo "${GRAFANA_TAG}" | cut -d "v" -f 2)
else
  echo "Provided tag is not a version tag, skipping packages release..."
	exit
fi

echo "Storing NPM artifacts"
ZIPFILE=grafana-npm-${_grafana_version}.tgz
tar -czf "$ZIPFILE" packages/*/dist packages/*/compiled
gsutil cp "$ZIPFILE" "gs://grafana-prerelease/artifacts/npm/$ZIPFILE"
echo "Done."