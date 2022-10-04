// Copyright 2022 MongoDB Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/evergreen-ci/shrub"
)

const (
	atlascli = "atlascli"
	mongocli = "mongocli"
	runOn    = "ubuntu1804-small"
	arm64    = "arm64"
	x86_64   = "x86_64"
	deb      = "deb"
	rpm      = "rpm"
)

var (
	serverVersions = []string{
		"4.2",
		"4.4",
		"5.0",
		"6.0",
	}
	oses = []string{
		"amazonlinux2",
		"centos7",
		"centos8",
		"rhel9",
		"debian10",
		"debian11",
		"ubuntu18.04",
		"ubuntu20.04",
		"ubuntu22.04",
	}
	repos      = []string{"org", "enterprise"}
	postPkgImg = map[string]string{
		"centos7":      "centos7-rpm",
		"centos8":      "centos8-rpm",
		"rhel9":        "rhel9-rpm",
		"amazonlinux2": "amazonlinux2-rpm",
		"ubuntu18.04":  "ubuntu18.04-deb",
		"ubuntu20.04":  "ubuntu20.04-deb",
		"ubuntu22.04":  "ubuntu22.04-deb",
		"debian10":     "debian10-deb",
		"debian11":     "debian11-deb",
	}
	newOs = map[string]string{
		"centos7":      "rhel70",
		"centos8":      "rhel80",
		"rhel9":        "rhel90",
		"amazonlinux2": "amazon2",
		"ubuntu18.04":  "ubuntu1804",
		"ubuntu20.04":  "ubuntu2004",
		"ubuntu22.04":  "ubuntu2204",
		"debian10":     "debian10",
		"debian11":     "debian11",
	}
	// if updating this list verify build/ci/repo_config.yaml matches.
	distros = map[string]Platform{
		"amazon2": {
			extension:     rpm,
			architectures: []string{x86_64, arm64},
		},
		"rhel70": {
			extension:     rpm,
			architectures: []string{x86_64},
		},
		"rhel80": {
			extension:     rpm,
			architectures: []string{x86_64, arm64},
		},
		"rhel90": {
			extension:     rpm,
			architectures: []string{x86_64, arm64},
		},
		"debian10": {
			extension:     rpm,
			architectures: []string{x86_64},
		},
		"debian11": {
			extension:     deb,
			architectures: []string{x86_64},
		},
		"ubuntu1804": {
			extension:     deb,
			architectures: []string{x86_64, arm64},
		},
		"ubuntu2004": {
			extension:     deb,
			architectures: []string{x86_64, arm64},
		},
		"ubuntu2204": {
			extension:     deb,
			architectures: []string{x86_64, arm64},
		},
	}
)

func buildDependency(toolName, os, serverVersion, repo string) shrub.TaskDependency {
	return shrub.TaskDependency{
		Name:    fmt.Sprintf("push_%s_%s_%s_stable", toolName, newOs[os], repo),
		Variant: fmt.Sprintf("release_%s_publish_%s", toolName, strings.ReplaceAll(serverVersion, ".", "")),
	}
}

func generateRepoTasks(c *shrub.Configuration, toolName string) {
	for _, serverVersion := range serverVersions {
		v := &shrub.Variant{
			BuildName:        fmt.Sprintf("test_repo_%v_%v", toolName, serverVersion),
			BuildDisplayName: fmt.Sprintf("Test %v on repo %v", toolName, serverVersion),
			DistroRunOn:      []string{runOn},
		}

		pkg := "mongodb-atlas-cli"
		entrypoint := "atlas"
		if toolName == mongocli {
			pkg = mongocli
			entrypoint = mongocli
		}

		for _, os := range oses {
			for _, repo := range repos {
				mongoRepo := "https://repo.mongodb.com"
				if repo == "org" {
					mongoRepo = "https://repo.mongodb.org"
				}

				t := &shrub.Task{
					Name: fmt.Sprintf("test_repo_%v_%v_%v_%v", toolName, os, repo, serverVersion),
				}
				t = t.Stepback(false).
					GitTagOnly(true).
					Dependency(buildDependency(toolName, os, serverVersion, repo)).
					Function("clone").
					FunctionWithVars("docker build repo", map[string]string{
						"server_version": serverVersion,
						"package":        pkg,
						"entrypoint":     entrypoint,
						"image":          os,
						"mongo_package":  fmt.Sprintf("mongodb-%v", repo),
						"mongo_repo":     mongoRepo,
					})
				c.Tasks = append(c.Tasks, t)
				v.AddTasks(t.Name)
			}
		}

		c.Variants = append(c.Variants, v)
	}
}

func generatePostPkgTasks(c *shrub.Configuration, toolName string) {
	v := &shrub.Variant{
		BuildName:        fmt.Sprintf("pkg_smoke_tests_docker_%v_generated", toolName),
		BuildDisplayName: fmt.Sprintf("Generated post packaging smoke tests (Docker / %v)", toolName),
		DistroRunOn:      []string{runOn},
	}

	for _, os := range oses {
		t := &shrub.Task{
			Name: fmt.Sprintf("pkg_test_%v_docker_%v", toolName, os),
		}
		t = t.Dependency(shrub.TaskDependency{
			Name:    "package_goreleaser",
			Variant: fmt.Sprintf("goreleaser_%v_snapshot", toolName),
		}).Function("clone").FunctionWithVars("docker build", map[string]string{
			"tool_name": toolName,
			"image":     postPkgImg[os],
		})
		c.Tasks = append(c.Tasks, t)
		v.AddTasks(t.Name)
	}

	c.Variants = append(c.Variants, v)
}

func generatePostPkgMetaTasks(c *shrub.Configuration, toolName string) {
	if toolName != atlascli {
		return
	}

	v := &shrub.Variant{
		BuildName:        fmt.Sprintf("pkg_smoke_tests_docker_meta_%s_generated", toolName),
		BuildDisplayName: fmt.Sprintf("Generated post packaging smoke tests (Meta / %s)", toolName),
		DistroRunOn:      []string{runOn},
	}

	for _, os := range oses {
		t := &shrub.Task{
			Name: fmt.Sprintf("pkg_test_%s_meta_docker_%s", toolName, os),
		}
		t = t.Dependency(shrub.TaskDependency{
			Name:    "package_goreleaser",
			Variant: fmt.Sprintf("goreleaser_%v_snapshot", toolName),
		}).Function("clone").
			FunctionWithVars("docker build meta", map[string]string{
				"tool_name": toolName,
				"image":     postPkgImg[os],
			})
		c.Tasks = append(c.Tasks, t)
		v.AddTasks(t.Name)
	}

	c.Variants = append(c.Variants, v)
}

type Platform struct {
	extension     string
	architectures []string
}

func generatePublishStableTasks(c *shrub.Configuration, toolName string) {
	dependency := []shrub.TaskDependency{
		{
			Name:    fmt.Sprintf("compile_%s", toolName),
			Variant: "code_health",
		},
		{
			Name:    fmt.Sprintf("release_%s", toolName),
			Variant: fmt.Sprintf("release_%s_github", toolName),
		},
	}
	for _, sv := range serverVersions {
		v := &shrub.Variant{
			BuildName:        fmt.Sprintf("generated_release_%s_publish_%s", toolName, sv),
			BuildDisplayName: fmt.Sprintf("Publish %s yum/apt %s", toolName, sv),
			DistroRunOn:      []string{"rhel80-small"},
		}
		publishVariant(
			c,
			v,
			toolName,
			sv,
			"_stable",
			dependency,
			true,
		)
	}
}

func generatePublishSnapshotTasks(c *shrub.Configuration, toolName string) {
	dependency := []shrub.TaskDependency{
		{
			Name:    fmt.Sprintf("compile_%s", toolName),
			Variant: "code_health",
		},
		{
			Name:    "package_goreleaser",
			Variant: fmt.Sprintf("goreleaser_%s_snapshot", toolName),
		},
	}
	publishVariant(
		c,
		c.Variant(fmt.Sprintf("publish_%s_snapshot", toolName)),
		toolName,
		"4.4",
		"",
		dependency,
		false,
	)
}

func publishVariant(c *shrub.Configuration, v *shrub.Variant, toolName, sv, stableSuffix string, dependency []shrub.TaskDependency, stable bool) {
	taskServerVersion := fmt.Sprintf("%s.0", sv)
	notaryKey := fmt.Sprintf("server-%s", sv)
	taskSv := "_" + sv
	if !stable {
		taskServerVersion = "4.4.0-rc3"
		notaryKey = "server-4.0"
		taskSv = ""
	}
	for k, d := range distros {
		for _, r := range repos {
			for _, a := range d.architectures {
				taskName := fmt.Sprintf("push_%s_%s_%s_%s%s%s", toolName, k, r, a, taskSv, stableSuffix)
				t := newPublishTask(taskName, toolName, d.extension, r, k, taskServerVersion, notaryKey, a, stable, dependency)
				c.Tasks = append(c.Tasks, t)
				v.AddTasks(t.Name)
			}
		}
	}
	c.Variants = append(c.Variants, v)
}

func newPublishTask(taskName, toolName, extension, edition, distro, taskServerVersion, notaryKey, arch string, stable bool, dependency []shrub.TaskDependency) *shrub.Task {
	t := &shrub.Task{
		Name: taskName,
	}
	t.Stepback(false).
		GitTagOnly(stable).
		Dependency(dependency...).
		Function("clone").
		Function("install curator").
		Patchable(false).
		FunctionWithVars("push", map[string]string{
			"tool_name":       toolName,
			"distro":          distro,
			"ext":             extension,
			"server_version":  taskServerVersion,
			"notary_key_name": notaryKey,
			"arch":            arch,
			"edition":         edition,
		})
	return t
}

var (
	ErrMissingOption = errors.New("missing option")
)

func run() error {
	var toolName, taskType string

	flag.StringVar(&taskType, "tasks", "", "type of task to be generated")
	flag.StringVar(&toolName, "tool_name", "", fmt.Sprintf("Tool to generate tasks for (%s or %s)", atlascli, mongocli))

	flag.Parse()

	if toolName == "" {
		return fmt.Errorf("%w: %s", ErrMissingOption, "tool_name")
	}

	if toolName != atlascli && toolName != mongocli {
		return fmt.Errorf("-tool_name must be either %q or %q", atlascli, mongocli)
	}

	if taskType == "" {
		return fmt.Errorf("%w: %s", ErrMissingOption, "tasks")
	}

	c := &shrub.Configuration{}

	switch taskType {
	case "repo":
		generateRepoTasks(c, toolName)
	case "postpkg":
		generatePostPkgTasks(c, toolName)
		generatePostPkgMetaTasks(c, toolName)
	case "snapshot":
		generatePublishSnapshotTasks(c, toolName)
	case "publish":
		generatePublishStableTasks(c, toolName)
	default:
		return errors.New("-tasks is invalid")
	}

	var b []byte
	b, err := json.MarshalIndent(c, "", "\t")
	if err != nil {
		return err
	}

	fmt.Println(string(b))

	return nil
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}
