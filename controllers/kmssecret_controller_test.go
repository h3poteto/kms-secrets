package controllers

import (
	"testing"
)

func TestShasumData(t *testing.T) {
	expected := "b6b66b55b6b03c6ee6abc0027095d38a35937eb3e6ff2dc9f2aafa846c704e3b"
	data := map[string][]byte{
		"API_KEY":  []byte("hoge"),
		"PASSWORD": []byte("fuga"),
	}
	sum := shasumData(data)
	if sum != expected {
		t.Errorf("shasum is not matched, expected: %s, returned: %s", expected, sum)
	}
}

func TestYamlParse(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "--- apikey",
			expected: "apikey",
		},
		{
			input:    "apikey",
			expected: "apikey",
		},
	}
CASE:
	for _, c := range cases {
		input := []byte(c.input)
		result, err := yamlParse(input)

		if err != nil {
			t.Error(err)
			continue CASE
		}
		if string(result) != c.expected {
			t.Errorf("Parsed result is not matched, expected: %s, result: %s", c.expected, result)
		}
	}
}
