package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestMakeTestTargets(t *testing.T) {
	makefile, err := os.ReadFile("../GNUmakefile")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name string
		args []string
		want string
	}{
		{
			name: "acceptance forwards test selection and options",
			args: []string{"testacc", "PRODUCT=uhost", "TESTARGS=-run '^TestAccSelected$$' -count=1"},
			want: "TF_ACC= test ./products/uhost -list ^TestAcc\nTF_ACC=1 test -cover ./products/uhost -v -run ^TestAccSelected$ -count=1 -timeout 120m -parallel=32\n",
		},
		{
			name: "sweep defaults to the package with registered sweepers",
			args: []string{"sweep"},
			want: "TF_ACC= test ./products/udpn -v -sweep=cn-bj2,cn-sh2\n",
		},
		{
			name: "sweep preserves explicit package region and filter",
			args: []string{"sweep", "TEST=./products/ulb", "SWEEP=cn-bj2", "SWEEPARGS=-sweep-run=selected"},
			want: "TF_ACC= test ./products/ulb -v -sweep=cn-bj2 -sweep-run=selected\n",
		},
		{
			name: "unit tests retain all packages",
			args: []string{"test"},
			want: "TF_ACC= test ./... -timeout=30s -parallel=32\n",
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.MkdirAll(filepath.Join(dir, "products", "uhost"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "GNUmakefile"), makefile, 0644); err != nil {
				t.Fatal(err)
			}
			// Execute the real Make recipes without permitting any Go or cloud work.
			fakeGo := `#!/bin/sh
printf 'TF_ACC=%s %s\n' "${TF_ACC:-}" "$*" >>"${MAKE_TEST_CALLS}"
case " $* " in
  *' -list '*) printf 'TestAccSelected\n' ;;
esac
`
			if err := os.WriteFile(filepath.Join(dir, "go"), []byte(fakeGo), 0755); err != nil {
				t.Fatal(err)
			}
			calls := filepath.Join(dir, "calls")
			command := exec.Command("make", append([]string{"--no-print-directory", "-o", "fmtcheck"}, test.args...)...)
			command.Dir = dir
			for _, variable := range os.Environ() {
				key := strings.SplitN(variable, "=", 2)[0]
				switch key {
				case "TF_ACC", "MAKEFLAGS", "MFLAGS", "MAKEOVERRIDES", "TEST", "TESTARGS", "PRODUCT", "SWEEP", "SWEEPARGS":
					continue
				}
				command.Env = append(command.Env, variable)
			}
			command.Env = append(command.Env,
				"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"MAKE_TEST_CALLS="+calls,
			)
			if output, err := command.CombinedOutput(); err != nil {
				t.Fatalf("make failed: %v\n%s", err, output)
			}
			got, err := os.ReadFile(calls)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != test.want {
				t.Fatalf("go invocations = %q, want %q", got, test.want)
			}
		})
	}
}
