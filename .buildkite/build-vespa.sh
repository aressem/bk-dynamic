#!/bin/bash

set -euo pipefail

ARCH=$1

export VESPA_MAVEN_EXTRA_OPTS="--show-version --batch-mode --no-snapshot-updates -Dmaven.javadoc.skip=true \
  -Dmaven.source.skip=true -DaltDeploymentRepository=local-repo::default::file:$(pwd)/artifacts/$ARCH/maven-repo" # deploy" # -Dmaven.repo.local=/tmp/vespa/mvnrepo
#export CCACHE_TMP_DIR="/tmp/ccache_tmp"
#export CCACHE_DATA_DIR="/tmp/vespa/ccache"
#export MAIN_CACHE_FILE="/tmp/vespa.tar"
export GOPATH="/root/.go"

export FACTORY_VESPA_VERSION=$VESPA_VERSION

NUM_THREADS=$NUM_CPU_LIMIT

VESPA_CMAKE_SANITIZERS_OPTION=""

export WORKDIR=/tmp

source /etc/profile.d/enable-gcc-toolset.sh

mkdir -p $(pwd)/artifacts/$ARCH/{rpms,maven-repo}

screwdriver/replace-vespa-version-in-poms.sh $VESPA_VERSION $(pwd)
time make -C client/go BIN=$WORKDIR/vespa-install/opt/vespa/bin SHARE=$WORKDIR/vespa-install/usr/share install-all

# Having both install and deploy in the same maven command does not work
sed -i 's,clean install,clean deploy,' bootstrap.sh

time ./bootstrap.sh full

# To allow Java and C++ tests to run in parallel, we need to copy the test jars
export VESPA_CPP_TEST_JARS=/tmp/VESPA_CPP_TEST_JARS
mkdir -p $VESPA_CPP_TEST_JARS
find . -type d -name target -exec find {} -mindepth 1 -maxdepth 1 -name *.jar \; | xargs -I '{}' cp '{}' $VESPA_CPP_TEST_JARS


time ./mvnw -T $NUM_THREADS $VESPA_MAVEN_EXTRA_OPTS deploy &> maven_output.log &

cmake3 -DVESPA_UNPRIVILEGED=no $VESPA_CMAKE_SANITIZERS_OPTION .
time make -j ${NUM_THREADS}
time ctest3 --output-on-failure -j ${NUM_THREADS}

echo "Waiting for Java build ..."
time wait || (cat maven_output.log && false)
cat maven_output.log

time make -j ${NUM_THREADS} install DESTDIR=$WORKDIR/vespa-install

# Build RPMS
ulimit -c 0
echo "%_binary_payload w10T8.zstdio" >> $HOME/.rpmmacros
time make  -f .copr/Makefile srpm outdir=$WORKDIR

time rpmbuild --rebuild --define="_topdir $WORKDIR/vespa-rpmbuild" \
                        --define "_debugsource_template %{nil}" \
                        --define "installdir $WORKDIR/vespa-install" $WORKDIR/*.src.rpm

mv /tmp/vespa-rpmbuild/RPMS/*/*.rpm $(pwd)/artifacts/$ARCH/rpms

# Sign artifacts
# Debug
#find $(pwd)/artifacts/maven-repo
#find $(pwd)/artifacts/rpms