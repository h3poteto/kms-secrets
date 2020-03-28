package controllers

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("shasumData", func() {
	It("should be matched", func() {
		data := map[string][]byte{
			"API_KEY":  []byte("hoge"),
			"PASSWORD": []byte("fuga"),
		}
		sum := shasumData(data)
		Expect(sum).To(Equal("b6b66b55b6b03c6ee6abc0027095d38a35937eb3e6ff2dc9f2aafa846c704e3b"))
	})
})

var _ = Describe("yamlParse", func() {
	It("yaml string should be parsed", func() {
		input := []byte("--- apikey")
		result, err := yamlParse(input)
		Expect(err).Should(BeNil())
		Expect(string(result)).To(Equal("apikey"))
	})
	It("string should not be changed", func() {
		input := []byte("apikey")
		result, err := yamlParse(input)
		Expect(err).Should(BeNil())
		Expect(string(result)).To(Equal("apikey"))
	})
})
