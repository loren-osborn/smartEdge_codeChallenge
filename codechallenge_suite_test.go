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

var JsonValidationSchemaPath string = "testdata/valid_output_schema.json"

// var _ = Describe("Code Challenge Program", func() {

// 	It("can be loaded from JSON", func() {
// 		Expect(book.Title).To(Equal("Les Miserables"))
// 		Expect(book.Author).To(Equal("Victor Hugo"))
// 		Expect(book.Pages).To(Equal(1488))
// 	})

// 	It("can extract the author's last name", func() {
// 		Expect(book.AuthorLastName()).To(Equal("Hugo"))
// 	})
// })
