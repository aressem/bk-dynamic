#!/bin/bash

set -euo pipefail

export VESPA_MAVEN_EXTRA_OPTS="--show-version --batch-mode --no-snapshot-updates" # -Dmaven.repo.local=/tmp/vespa/mvnrepo
#export CCACHE_TMP_DIR="/tmp/ccache_tmp"
#export CCACHE_DATA_DIR="/tmp/vespa/ccache"
#export MAIN_CACHE_FILE="/tmp/vespa.tar"
#export GOPATH="/tmp/vespa/go"

export FACTORY_VESPA_VERSION=$VESPA_VERSION
NUM_THREADS=$(( $(nproc) + 2 ))

VESPA_CMAKE_SANITIZERS_OPTION=""

export WORKDIR=/tmp

source /etc/profile.d/enable-gcc-toolset.sh

screwdriver/replace-vespa-version-in-poms.sh $VESPA_VERSION $(pwd)
time make -C client/go BIN=$WORKDIR/vespa-install/opt/vespa/bin SHARE=$WORKDIR/vespa-install/usr/share install-all
time ./bootstrap.sh java
time ./mvnw -T $NUM_THREADS $VESPA_MAVEN_EXTRA_OPTS install
cmake3 -DVESPA_UNPRIVILEGED=no $VESPA_CMAKE_SANITIZERS_OPTION .
time make -j ${NUM_THREADS}
time ctest3 --output-on-failure -j ${NUM_THREADS}
time make -j ${NUM_THREADS} install DESTDIR=$WORKDIR/vespa-install
