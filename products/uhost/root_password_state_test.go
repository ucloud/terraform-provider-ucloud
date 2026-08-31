package uhost

import "testing"

func TestShouldPreserveInstanceRootPasswordState(t *testing.T) {
	tests := []struct {
		loginMode string
		want      bool
	}{
		{loginMode: "", want: true},
		{loginMode: "Password", want: true},
		{loginMode: "KeyPair"},
		{loginMode: "FutureLoginMode"},
	}

	for _, test := range tests {
		t.Run(test.loginMode, func(t *testing.T) {
			got := shouldPreserveInstanceRootPasswordState(test.loginMode)
			if got != test.want {
				t.Fatalf("shouldPreserveInstanceRootPasswordState(%q) = %v, want %v", test.loginMode, got, test.want)
			}
		})
	}
}
