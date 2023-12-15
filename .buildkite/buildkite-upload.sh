#!/bin/bash
 
set -eo pipefail
 
export BUILDKITE_PIPELINE_NO_INTERPOLATION=true
 
echo "--- :Buildkite: build information"
echo "BUILDKITE_BUILD_ID = $BUILDKITE_BUILD_ID"
 
 
echo "--- :pipeline_upload: uploading pipeline"

cat <<EOF | buildkite-agent pipeline upload 

steps:
  - label: ":pipeline: dynamicly generated build"
    plugins:
      - hasura/smooth-checkout#v3.1.0:
          repos:
            - config:
              - url: "https://github.com/vespa-engine/vespa.git"
              - ref: "master"
              - clone_flags: "--depth 1"
    command: find $HOME/.buildkite 
EOF

