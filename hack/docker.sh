#!/bin/bash

chart_appVersion=$(helm show chart chart | awk '/^appVersion:/ {print $2}')
chart_name=$(helm show chart chart | awk '/^name:/ {print $2}')

echo image name: $chart_name
echo image tag: $chart_appVersion

docker buildx build -f Dockerfile --platform linux/amd64 -t no8ge/$chart_name:$chart_appVersion . --push
