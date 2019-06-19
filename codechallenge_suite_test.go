package codechallenge_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	// "github.com/xeipuuv/gojsonschema"

	"testing"
)

func TestSmartEdgeCodingChallenge(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "SmartEdge: CodingChallenge Suite")
}

var _ = Describe("Code Challenge Program", func() {
	Describe("Invoking program", func() {
		Context("With input longer than 250 characters", func() {
		})
	})
	// Describe("Processing Input", func() {
	// 	Context("With input longer than 250 characters", func() {
	// 	})
	// })
	// It("can be loaded from JSON", func() {
	// 	Expect(book.Title).To(Equal("Les Miserables"))
	// 	Expect(book.Author).To(Equal("Victor Hugo"))
	// 	Expect(book.Pages).To(Equal(1488))
	// })

	// It("can extract the author's last name", func() {
	// 	Expect(book.AuthorLastName()).To(Equal("Hugo"))
	// })
})
