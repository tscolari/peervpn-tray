package peervpn_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPeervpn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Peervpn Suite")
}
