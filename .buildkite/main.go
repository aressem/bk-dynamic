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

	cmd := fmt.Sprintf("'"+
		"pwd "+
		//"&& du -sh /root/.* "+
		"&& ccache -s "+
		"&& ccache -z -M 20G "+
		"&& git clone --quiet --depth 1000 https://github.com/vespa-engine/vespa "+
		"&& (cd vespa && git checkout %s) && export VESPA_VERSION=%s "+
		"&& (cd vespa && git tag v\\$VESPA_VERSION) "+
		"&& echo make -C vespa -f .copr/Makefile rpms outdir=$(pwd) "+
		"&& ccache -s "+
		"&& buildkite-agent artifact upload vespa/README.md "+
		"&& buildkite-agent artifact upload vespa/README.md s3://381492154096-build-artifacts/\\$BUILDKITE_JOB_ID' ",
		vespaGitref, vespaVersion)

	//fmt.Println(cmd)

	pipe := &pipeline.Pipeline{}
	step := &pipeline.CommandStep{
		Label: "Vespa build!",
	}
	plug := &pipeline.Plugin{
		Source: "kubernetes",
		Config: map[string]any{
			"podSpec": map[string]any{
				/*				"volumes": []any{
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
							},*/
				"containers": []any{
					map[string]any{
						"args": []string{
							//"'pwd && git clone --quiet --depth 1000 https://github.com/vespa-engine/vespa && buildkite-agent artifact upload vespa/README.md && export VESPA_VERSION=8.999.1 && (cd vespa && git tag v\\$VESPA_VERSION) && make -C vespa -f .copr/Makefile rpms outdir=$(pwd) && buildkite-agent artifact upload vespa-8.999.1-1.el8.src.rpm'",
							cmd,
						},
						"command": []any{
							"sh",
							"-c",
						},
						"image": "docker.io/vespaengine/vespa-build-almalinux-8",
						"resources": map[string]any{
							"limits": map[string]any{
								"cpu":    "15",
								"memory": "30G",
							},
						},
						/*						"volumeMounts": []any{
												map[string]any{
													"mountPath": "/root/.m2",
													"name":      "vespa-build-maven-cache",
												},
												map[string]any{
													"mountPath": "/root/.ccache",
													"name":      "vespa-build-ccache",
												},
											},*/
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
