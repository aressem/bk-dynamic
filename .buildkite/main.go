package main

import (
	"fmt"
	"github.com/buildkite/go-pipeline"
	"gopkg.in/yaml.v3"
	"os"
)

func main() {

	vespaVersion := os.Getenv("VESPA_VERSION")
	if len(vespaVersion) == 0 {
		vespaVersion = "8.999.1"
	}
	vespaGitref := os.Getenv("VESPA_GITREF")
	if len(vespaGitref) == 0 {
		vespaGitref = "0420245d142858687f6d8b412fd08e8edf1018d5"
	}
	//	fmt.Printf("VESPA_VERSION: %s\n\n", vespaVersion)
	//	fmt.Printf("VESPA_GITREF: %s\n\n", vespaGitref)

	cmd := fmt.Sprintf("'" +
		"pwd " +
		//"&& du -sh /root/.* "+
		"&& mkdir -p /tmp/ccache_tmp " +
		"&& ccache -s -p" +
		"&& ccache -z -o temporary_dir=/tmp/ccache_tmp -o compression=true -M 20G " +
		"&& export CCACHE_COMPRESS=1 " +
		"&& export FACTORY_VESPA_VERSION=\\$VESPA_VERSION " +
		"&& (git tag v\\$VESPA_VERSION) " +
		"&& git clone --depth 1 https://github.com/aressem/bk-dynamic" +
		"&& bk-dynamic/build-vespa.sh " +
		"&& ccache -s " +
		"&& buildkite-agent artifact upload README.md " +
		"&& buildkite-agent artifact upload README.md s3://381492154096-build-artifacts/\\$BUILDKITE_JOB_ID " +
		"&& tar -C /root --exclude '.m2/repository/com/yahoo/vespa' -cvf cache.tar  .ccache .m2/repository " +
		"&& du -sh /root/.m2 && du -sh /root/.ccache " +
		"&& buildkite-agent artifact upload cache.tar s3://381492154096-build-artifacts " +
		"'")

	//fmt.Println(cmd)

	pipe := &pipeline.Pipeline{}
	step := &pipeline.CommandStep{
		Label: "Vespa build!",
		Env: map[string]string{
			"BUILDKITE_GIT_CLONE_FLAGS": "--depth 1000",
		},
	}

	plug := &pipeline.Plugin{
		Source: "kubernetes",
		Config: map[string]any{
			"podSpec": map[string]any{
				"volumes": []any{
					map[string]any{
						"name": "vespa-build-maven-cache",
						"persistentVolumeClaim": map[string]string{
							"claimName": "vespa-build-maven-cache",
						},
					},
					map[string]any{
						"name": "vespa-build-ccache",
						"persistentVolumeClaim": map[string]string{
							"claimName": "vespa-build-ccache",
						},
					},
				},
				"containers": []any{
					map[string]any{
						"args": []string{
							cmd,
						},
						"command": []any{
							"sh",
							"-c",
						},
						"env": []any{
							map[string]string{
								"name":  "BUILDKITE_S3_ACL",
								"value": "private",
							},
						},
						"image": "docker.io/vespaengine/vespa-build-almalinux-8",
						"resources": map[string]any{
							"limits": map[string]any{
								"cpu":    "15",
								"memory": "30G",
							},
						},
						"volumeMounts": []any{
							map[string]any{
								"mountPath": "/root/.m2",
								"name":      "vespa-build-maven-cache",
							},
							map[string]any{
								"mountPath": "/root/.ccache",
								"name":      "vespa-build-ccache",
							},
						},
					},
				},
				"nodeSelector": map[string]any{
					"node-arch": "arm64",
				},
				"priorityClassName": "system-node-critical",
			},
		},
	}

	step.Plugins = append(step.Plugins, plug)
	pipe.Steps = append(pipe.Steps, step)

	yout, err := yaml.Marshal(pipe)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(yout))
}
