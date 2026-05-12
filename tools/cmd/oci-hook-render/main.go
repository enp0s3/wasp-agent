//Licensed under the Apache License, Version 2.0 (the "License");
//you may not use this file except in compliance with the License.
//You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
//Unless required by applicable law or agreed to in writing, software
//distributed under the License is distributed on an "AS IS" BASIS,
//WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//See the License for the specific language governing permissions and
//limitations under the License.

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/openshift-virtualization/wasp-agent/pkg/wasp/config"
	oci_hook_render "github.com/openshift-virtualization/wasp-agent/pkg/wasp/oci-hook-render"
)

const (
	defaultPodName        = "wasp-agent-preview"
	defaultHookTemplate   = "OCI-hook/hookscript.template"
	defaultOutputPath     = "oci-hook-swap.sh"
)

func main() {
	var crioConfigDir string
	var outputPath string
	var hookTemplate string

	flag.StringVar(&crioConfigDir, "crio-config-dir", "", "path to the CRI-O configuration drop-in directory (required)")
	flag.StringVar(&outputPath, "o", defaultOutputPath, "output path for the rendered OCI hook script")
	flag.StringVar(&hookTemplate, "template", defaultHookTemplate, "path to the OCI hook script template")
	flag.Parse()

	if crioConfigDir == "" {
		fmt.Fprintf(os.Stderr, "Error: -crio-config-dir is required\n\n")
		flag.Usage()
		os.Exit(1)
	}

	crioConfig := config.New("", crioConfigDir)
	runtime, err := crioConfig.GetRuntime()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error detecting runtime from CRI-O config: %v\n", err)
		os.Exit(1)
	}
	fmt.Fprintf(os.Stderr, "Detected runtime: %s\n", runtime)

	renderer := oci_hook_render.New(hookTemplate, outputPath, runtime)
	if err := renderer.Render(); err != nil {
		fmt.Fprintf(os.Stderr, "Error rendering hook script: %v\n", err)
		os.Exit(1)
	}

	if err := os.Chmod(outputPath, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error setting file permissions: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "OCI hook script rendered to %s (pod suffix: %s)\n", outputPath, defaultPodName)
}
