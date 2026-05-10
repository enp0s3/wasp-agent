package consts

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Hook path generation", func() {
	const testSuffix = "wasp-agent-abc12"

	It("should generate the correct hook script path with pod name suffix", func() {
		Expect(HookScriptPath(testSuffix)).To(Equal("/host/opt/oci-hook-swap-wasp-agent-abc12.sh"))
	})

	It("should generate the correct hook config path with pod name suffix", func() {
		Expect(HookConfigPath(testSuffix)).To(Equal("/host/run/containers/oci/hooks.d/swap-for-burstable-wasp-agent-abc12.json"))
	})

	It("should produce unique paths for different suffixes", func() {
		Expect(HookScriptPath("wasp-agent-abc12")).ToNot(Equal(HookScriptPath("wasp-agent-def34")))
		Expect(HookConfigPath("wasp-agent-abc12")).ToNot(Equal(HookConfigPath("wasp-agent-def34")))
	})
})
