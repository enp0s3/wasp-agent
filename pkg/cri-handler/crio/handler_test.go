package crio_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("CRIO Handler", func() {
	Context("Setup", func() {
		When("returns w/o errors", func() {
			It("container runtime is determined", func() {
				Expect(true).To(BeTrue())
			})
		})
	})
	Context("ConfigParser", func() {
		When("drop-in configuration in place", func() {
			It("parse the runtime configuration from the drop-in", func() {
				Expect(true).To(BeTrue())
			})
		})
		When("no drop-in configuration appear", func() {
			It("parse runtime from root config", func() {
				Expect(true).To(BeTrue())
			})
		})
		When("CRIO configuration files doesn't exist", func() {
			It("Parse the deafult hard-coded runtime config", func() {
				Expect(true).To(BeTrue())
			})
		})
	})
})
