package crio_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCrio(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Crio Suite")
}
