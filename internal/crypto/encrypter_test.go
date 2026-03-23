package crypto

import "testing"

func TestIsEncrypted(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{name: "age ciphertext", value: "age1qyqszqgpqyqszqgpqyqszqgp", want: true},
		{name: "age prefix only", value: "age1", want: true},
		{name: "plain string", value: "hello-world", want: false},
		{name: "empty string", value: "", want: false},
		{name: "age without 1", value: "ageXYZ", want: false},
		{name: "starts with age1 in middle", value: "foo-age1bar", want: false},
		{name: "AGE1 uppercase", value: "AGE1abc", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEncrypted(tt.value)
			if got != tt.want {
				t.Errorf("IsEncrypted(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
