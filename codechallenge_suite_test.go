package codechallenge_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSmartEdgeCodingChallenge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmartEdge: CodingChallenge Suite")
}
