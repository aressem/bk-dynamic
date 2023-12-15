#!/bin/bash
 
set -eo pipefail
 
export BUILDKITE_PIPELINE_NO_INTERPOLATION=true
 
echo "--- :Buildkite: build information"
echo "BUILDKITE_BUILD_ID = $BUILDKITE_BUILD_ID"
 
 
echo "--- :pipeline_upload: uploading pipeline"

cat <<EOF | /usr/bin/buildkite-agent pipeline upload 

steps:
  - label: ":pipeline: dynamicly generated pipeline"
    command: find / 
EOF

