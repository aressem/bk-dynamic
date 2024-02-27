package main

import (
    "fmt"
    "github.com/buildkite/go-pipeline"
    "gopkg.in/yaml.v3"
)

func main() {

    pipe := &pipeline.Pipeline{}
    step := &pipeline.CommandStep{
        Label: "Vespa build!",
    }

    plug := &pipeline.Plugin{
        Source: "kubernetes",
        Config: map[string]any{
            "podSpec": map[string]any{
                "volumes": []any{
                    map[string]any{
                        "name": "maven-m2-cache",
                        "persistentVolumeClaim": map[string]string{
                            "claimName": "vespa-build-maven-cache",
                        },
                    },
                    map[string]any{
                        "name": "gcc-ccache",
                        "persistentVolumeClaim": map[string]string{
                            "claimName": "vespa-build-ccache",
                        },
                    },
                },
                "containers": []any{
                    map[string]any{
                        "args": []string{
                            "'pwd && git clone --quiet --depth 1 https://github.com/vespa-engine/vespa && buildkite-agent artifact upload vespa/README.md && export VESPA_VERSION=8.999.1 && (cd vespa && git tag v\\$VESPA_VERSION) && make -C vespa -f .copr/Makefile rpms outdir=$(pwd) && buildkite-agent artifact upload vespa-8.999.1-1.el8.src.rpm'",
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
