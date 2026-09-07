package scripts_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecheckProductOwnership(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq is required by the GitHub runner script")
	}
	script, err := filepath.Abs("recheck-product-ownership.sh")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name        string
		apiFailure  bool
		mode        string
		number      string
		wantFailure bool
		wantCalls   string
		wantOutput  string
	}{
		{name: "checks all proposals despite a rejected proposal", mode: "check", wantFailure: true, wantCalls: "1\n3"},
		{name: "API failure is not a successful check", mode: "check", apiFailure: true, wantFailure: true},
		{name: "lists only non-Core ownership proposals", mode: "list", wantOutput: "1\n3"},
		{name: "checks one PR for the workflow matrix", mode: "check", number: "3", wantCalls: "3"},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, ".github"), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, ".github", "product-owners.json"), []byte(`{"core":{"github_users":["CoreOwner"]}}`), 0644); err != nil {
				t.Fatal(err)
			}
			writeExecutable := func(name, contents string) string {
				t.Helper()
				filename := filepath.Join(dir, name)
				if err := os.WriteFile(filename, []byte(contents), 0755); err != nil {
					t.Fatal(err)
				}
				return filename
			}
			checker := writeExecutable("checker", `#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == '-config' && "$3" == '-event' ]]
jq -e '.repository.full_name == "ucloud/test" and .sender.login == .pull_request.user.login' "$4" >/dev/null
number="$(jq -r '.pull_request.number' "$4")"
printf '%s\n' "$number" >>"${REVIEW_CALLS}"
[[ "$number" == '3' ]]
`)
			writeExecutable("go", `#!/usr/bin/env bash
set -euo pipefail
[[ "$1" == 'build' && "$2" == '-o' && "$4" == './cmd/product-ownership' ]]
cp "${REVIEW_CHECKER}" "$3"
`)
			writeExecutable("gh", `#!/usr/bin/env bash
set -euo pipefail
if [[ "${REVIEW_API_FAILURE}" == 'true' ]]; then exit 1; fi
case "$*" in
  'api --paginate repos/ucloud/test/pulls?state=open&base=master&per_page=100 --jq .[].number')
    printf '1\n2\n3\n4\n5\n' ;;
  'api repos/ucloud/test/pulls/'[12345])
    number="${2##*/}"
    changed=1
    if [[ "$number" == '4' ]]; then changed=2; fi
    author=NewOwner
    if [[ "$number" == '5' ]]; then author=COREOWNER; fi
    jq -n --argjson number "$number" --argjson changed "$changed" --arg author "$author" \
      '{number:$number,state:"open",base:{ref:"master"},changed_files:$changed,user:{login:$author}}' ;;
  'api repos/ucloud/test/pulls/'[13]'/files --jq .[].filename')
    printf '.github/product-owners.json\n' ;;
  'api repos/ucloud/test/pulls/2/files --jq .[].filename')
    printf 'products/ulb/product.go\n' ;;
  *) printf 'unexpected request: %s\n' "$*" >&2; exit 2 ;;
esac
`)
			calls := filepath.Join(dir, "calls")
			command := exec.Command("bash", script, test.mode)
			command.Dir = dir
			failure := "false"
			if test.apiFailure {
				failure = "true"
			}
			command.Env = append(os.Environ(),
				"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GITHUB_REPOSITORY=ucloud/test", "RUNNER_TEMP="+dir,
				"PRODUCT_OWNERSHIP_PR="+test.number,
				"REVIEW_CHECKER="+checker, "REVIEW_CALLS="+calls, "REVIEW_API_FAILURE="+failure,
			)
			output, err := command.CombinedOutput()
			if (err != nil) != test.wantFailure {
				t.Fatalf("recheck error = %v, wantFailure = %t; output: %s", err, test.wantFailure, output)
			}
			got, readErr := os.ReadFile(calls)
			if test.wantCalls == "" {
				if !os.IsNotExist(readErr) {
					t.Fatalf("checker should not have been invoked: %s, %v", got, readErr)
				}
			} else if readErr != nil || strings.TrimSpace(string(got)) != test.wantCalls {
				t.Fatalf("checked PRs = %q, error = %v; want %q; output: %s", got, readErr, test.wantCalls, output)
			}
			if test.wantOutput != "" && strings.TrimSpace(string(output)) != test.wantOutput {
				t.Fatalf("output = %q, want %q", output, test.wantOutput)
			}
			leftovers, err := filepath.Glob(filepath.Join(dir, "product-ownership-review.*"))
			if err != nil || len(leftovers) != 0 {
				t.Fatalf("review temp directory was not cleaned: %v, %v", leftovers, err)
			}
		})
	}
}
