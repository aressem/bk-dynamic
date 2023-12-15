#!/bin/bash
 
set -eo pipefail
 
export BUILDKITE_PIPELINE_NO_INTERPOLATION=true
 
echo "--- :Buildkite: build information"
echo "BUILDKITE_BUILD_ID = $BUILDKITE_BUILD_ID"
 
 
echo "--- :pipeline_upload: uploading pipeline"

cat <<EOF | buildkite-agent pipeline upload 

steps:
  - label: ":pipeline: dynamicly generated build"
    command: |
      git clone --depth 1 https://github.com/vespa-engine/vespa
      find $HOME/.buildkite | grep vespa
EOF

