package main

import (
	"fmt"
	"github.com/buildkite/go-pipeline"
	"gopkg.in/yaml.v3"
	"os"
)

func isPullRequest() bool {
	return os.Getenv("BUILDKITE_PULL_REQUEST") != "false"
}
func getVolumes(arch string) []any {
	if isPullRequest() {
		return []any{}
	} else {
		return []any{
			map[string]any{
				"name": "vespa-build-maven-cache",
				"persistentVolumeClaim": map[string]string{
					"claimName": "vespa-build-maven-cache-" + arch,
				},
			},
			map[string]any{
				"name": "vespa-build-ccache",
				"persistentVolumeClaim": map[string]string{
					"claimName": "vespa-build-ccache-" + arch,
				},
			},
		}
	}
}

func getVolumeMounts() []any {
	if isPullRequest() {
		return []any{}
	} else {
		return []any{
			map[string]any{
				"name":      "vespa-build-maven-cache",
				"mountPath": "/root/.m2",
			},
			map[string]any{
				"name":      "vespa-build-ccache",
				"mountPath": "/root/.ccache",
			},
		}
	}
}

func restoreCache(arch string) string {
	if isPullRequest() {
		return "&& dnf install -y awscli " +
			"&& ( (aws s3 cp s3://381492154096-build-artifacts/cache-" + arch + ".tar - | tar -C /root -x) || true) "
	} else {
		return ""
	}
}

func saveCache(arch string) string {
	if isPullRequest() {
		return ""
	} else {
		return "&& mkdir -p /root/.ccache /root/.m2/repository /root/.go " +
			"&& tar -C /root --exclude '.m2/repository/com/yahoo/vespa' -cf cache-" + arch + ".tar  .ccache .m2/repository .go " +
			"&& buildkite-agent artifact upload cache-" + arch + ".tar s3://381492154096-build-artifacts "
	}
}

func saveArtifacts(version string, arch string) string {
	if isPullRequest() {
		return ""
	} else {
		return "&& buildkite-agent artifact upload \"artifacts/**/*\" s3://381492154096-build-artifacts/" + version + " "
	}
}

func getVespaVersion() string {
	vespaVersion := os.Getenv("VESPA_VERSION")
	if len(vespaVersion) == 0 {
		if isPullRequest() {
			vespaVersion = "8.999.999"
			os.Setenv("VESPA_VERSION", vespaVersion)
		} else {
			panic("VESPA_VERSION not set")
		}
	}
	return vespaVersion
}

func getBuildCommand(vespaVersion string, arch string) string {
	return fmt.Sprintf("'" +
		"pwd " +
		restoreCache(arch) +
		"&& mkdir -p /tmp/ccache_tmp " +
		"&& ccache -s -p" +
		"&& ccache -z -o temporary_dir=/tmp/ccache_tmp -o compression=true -M 20G " +
		"&& export CCACHE_COMPRESS=1 " +
		"&& export FACTORY_VESPA_VERSION=\\$VESPA_VERSION " +
		"&& (git tag v\\$VESPA_VERSION) " +
		"&& git clone --depth 1 https://github.com/aressem/bk-dynamic" +
		"&& bk-dynamic/.buildkite/build-vespa.sh " + arch + " " +
		"&& ccache -s " +
		"&& buildkite-agent artifact upload '*.log' " +
		"&& du -sh /root/.m2 && du -sh /root/.ccache " +
		saveCache(arch) +
		saveArtifacts(vespaVersion, arch) +
		"'")

}

func main() {

	vespaVersion := getVespaVersion()

	//fmt.Println(cmd)

	pipe := &pipeline.Pipeline{}

	// iterate over a list of strings
	for _, arch := range []string{"amd64", "arm64"} {

		step := &pipeline.CommandStep{
			Label: "Build Vespa version " + vespaVersion + " on linux/" + arch,
			Env: map[string]string{
				"BUILDKITE_GIT_CLONE_FLAGS": "--filter tree:0",
			},
		}

		plug := &pipeline.Plugin{
			Source: "kubernetes",
			Config: map[string]any{
				"gitEnvFrom": []any{
					map[string]any{
						"secretRef": map[string]string{
							"name": "github-vespaai-buildkite-access",
						},
					},
				},
				"podSpec": map[string]any{
					"volumes": getVolumes(arch),

					"containers": []any{
						map[string]any{
							"name": "build-container",
							"args": []string{
								getBuildCommand(vespaVersion, arch),
							},
							"command": []any{
								"sh",
								"-c",
							},
							"env": []any{
								map[string]string{
									"name":  "VESPA_VERSION",
									"value": vespaVersion,
								},
								map[string]string{
									"name":  "BUILDKITE_S3_ACL",
									"value": "private",
								},
								map[string]any{
									"name": "NUM_CPU_LIMIT",
									"valueFrom": map[string]any{
										"resourceFieldRef": map[string]string{
											"containerName": "build-container",
											"resource":      "limits.cpu",
										},
									},
								},
							},
							"image": "docker.io/vespaengine/vespa-build-almalinux-8",
							"resources": map[string]any{
								"limits": map[string]any{
									"cpu":    "15",
									"memory": "30G",
								},
							},
							"volumeMounts": getVolumeMounts(),
						},
					},
					"nodeSelector": map[string]any{
						"node-arch": arch,
					},
					"priorityClassName": "system-node-critical",
				},
			},
		}

		step.Plugins = append(step.Plugins, plug)
		pipe.Steps = append(pipe.Steps, step)
	}

	yout, err := yaml.Marshal(pipe)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(yout))
}
