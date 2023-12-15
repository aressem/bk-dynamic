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
      pwd
      git clone --quiet --depth 1 https://github.com/vespa-engine/vespa
      
      export VESPA_VERSION=8.999.1
      (cd vespa && git tag v${VESPA_VERSION})

      make -C vespa -f .copr/Makefile rpms outdir=$(pwd)

      buildkite-agent artifact upload *.rpm

EOF

