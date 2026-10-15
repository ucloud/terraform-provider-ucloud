package uhost

import "testing"

func TestDiffValidateInstanceNetCapability(t *testing.T) {
	tests := []struct {
		name    string
		old     string
		new     string
		wantErr bool
	}{
		{name: "empty old is ignored", old: "", new: "super"},
		{name: "empty new is ignored", old: "super", new: ""},
		{name: "normal to super upgrade", old: "normal", new: "super"},
		{name: "normal to ultra upgrade", old: "normal", new: "ultra"},
		{name: "super to normal downgrade", old: "super", new: "normal"},
		{name: "extreme to normal downgrade", old: "extreme", new: "normal"},
		{name: "unchanged value", old: "super", new: "super"},
		{name: "switching enhancement levels is rejected", old: "super", new: "ultra", wantErr: true},
		{name: "downgrade across enhancement levels is rejected", old: "extreme", new: "super", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := diffValidateInstanceNetCapability(test.old, test.new, nil)
			if (err != nil) != test.wantErr {
				t.Fatalf("diffValidateInstanceNetCapability(%q, %q) error = %v, wantErr %v", test.old, test.new, err, test.wantErr)
			}
		})
	}
}
