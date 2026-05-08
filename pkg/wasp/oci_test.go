package wasp

import (
	"encoding/json"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("OCI hook lifecycle", func() {
	var tmpDir string

	BeforeEach(func() {
		tmpDir = GinkgoT().TempDir()
	})

	Context("renderHookConfig", func() {
		var templatePath string

		BeforeEach(func() {
			templatePath = filepath.Join(tmpDir, "hook-config.template")
			templateContent := `{
  "version": "1.0.0",
  "hook": {
    "path": "{{ .HookScriptPath }}"
  },
  "when": {
    "always": true
  },
  "stages": [
    "poststart"
  ]
}`
			Expect(os.WriteFile(templatePath, []byte(templateContent), 0644)).To(Succeed())
		})

		It("should render the template with the correct script path stripped of /host prefix", func() {
			configPath := filepath.Join(tmpDir, "swap-for-burstable-wasp-agent-abc12.json")
			scriptPath := "/host/opt/oci-hook-swap-wasp-agent-abc12.sh"

			Expect(renderHookConfigFromTemplate(templatePath, configPath, scriptPath)).To(Succeed())

			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())

			var result map[string]interface{}
			Expect(json.Unmarshal(content, &result)).To(Succeed())

			hook := result["hook"].(map[string]interface{})
			Expect(hook["path"]).To(Equal("/opt/oci-hook-swap-wasp-agent-abc12.sh"))
			Expect(result["version"]).To(Equal("1.0.0"))

			when := result["when"].(map[string]interface{})
			Expect(when["always"]).To(BeTrue())

			stages := result["stages"].([]interface{})
			Expect(stages).To(HaveLen(1))
			Expect(stages[0]).To(Equal("poststart"))
		})

		It("should produce valid JSON output", func() {
			configPath := filepath.Join(tmpDir, "config.json")
			Expect(renderHookConfigFromTemplate(templatePath, configPath, "/host/opt/test-hook.sh")).To(Succeed())

			content, err := os.ReadFile(configPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(json.Valid(content)).To(BeTrue())
		})

		It("should fail when the template file does not exist", func() {
			configPath := filepath.Join(tmpDir, "config.json")
			err := renderHookConfigFromTemplate("/nonexistent/template", configPath, "/host/opt/test.sh")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("error while parsing hook config template"))
		})

		It("should fail when the output path is not writable", func() {
			configPath := "/nonexistent-dir/config.json"
			err := renderHookConfigFromTemplate(templatePath, configPath, "/host/opt/test.sh")
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("failed to create hook config"))
		})
	})

	Context("cleanupOCIHook", func() {
		It("should remove both hook files when they exist", func() {
			scriptFile := filepath.Join(tmpDir, "script.sh")
			configFile := filepath.Join(tmpDir, "config.json")
			Expect(os.WriteFile(scriptFile, []byte("#!/bin/bash"), 0755)).To(Succeed())
			Expect(os.WriteFile(configFile, []byte("{}"), 0644)).To(Succeed())

			cleanupFiles(configFile, scriptFile)

			Expect(scriptFile).ToNot(BeAnExistingFile())
			Expect(configFile).ToNot(BeAnExistingFile())
		})

		It("should not panic when files do not exist", func() {
			scriptFile := filepath.Join(tmpDir, "nonexistent-script.sh")
			configFile := filepath.Join(tmpDir, "nonexistent-config.json")

			Expect(func() {
				cleanupFiles(configFile, scriptFile)
			}).ToNot(Panic())
		})

		It("should remove one file even if the other does not exist", func() {
			scriptFile := filepath.Join(tmpDir, "script.sh")
			configFile := filepath.Join(tmpDir, "nonexistent-config.json")
			Expect(os.WriteFile(scriptFile, []byte("#!/bin/bash"), 0755)).To(Succeed())

			cleanupFiles(configFile, scriptFile)

			Expect(scriptFile).ToNot(BeAnExistingFile())
		})
	})
})
