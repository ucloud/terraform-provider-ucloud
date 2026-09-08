package scripts_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const (
	releaseHeadSHA  = "0123456789abcdef0123456789abcdef01234567"
	releaseMergeSHA = "89abcdef0123456789abcdef0123456789abcdef"
)

func TestReleasePolicyChecksPullRequestLabels(t *testing.T) {
	script := releasePolicyScript(t)
	for _, test := range []struct {
		name        string
		labels      []string
		wantFailure bool
		wantMessage string
		wantState   string
	}{
		{name: "one supported label", labels: []string{"release:minor", "uhost"}, wantState: "success"},
		{name: "missing release label", labels: []string{"uhost"}, wantFailure: true, wantMessage: "exactly one release label", wantState: "failure"},
		{name: "multiple release labels", labels: []string{"release:minor", "release:patch"}, wantFailure: true, wantMessage: "exactly one release label", wantState: "failure"},
		{name: "unsupported release label", labels: []string{"release:preview"}, wantFailure: true, wantMessage: "unsupported release label", wantState: "failure"},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			eventPath := filepath.Join(dir, "event.json")
			writeReleaseEvent(t, eventPath, test.labels)
			calls := filepath.Join(dir, "gh-calls")
			writeReleaseFakeGH(t, dir)

			command := exec.Command("bash", script, "check-pr")
			command.Env = append(os.Environ(),
				"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GITHUB_EVENT_PATH="+eventPath,
				"GITHUB_SERVER_URL=https://github.example",
				"GITHUB_RUN_ID=123",
				"RELEASE_GH_CALLS="+calls,
			)
			output, err := command.CombinedOutput()
			if (err != nil) != test.wantFailure {
				t.Fatalf("check-pr error = %v, wantFailure = %t; output: %s", err, test.wantFailure, output)
			}
			if test.wantMessage != "" && !strings.Contains(string(output), test.wantMessage) {
				t.Fatalf("check-pr output = %q, want %q", output, test.wantMessage)
			}
			contents, err := os.ReadFile(calls)
			if err != nil {
				t.Fatalf("read status calls: %v", err)
			}
			got := string(contents)
			if count := strings.Count(got, "state=pending"); count != 2 {
				t.Fatalf("pending status count = %d, want 2; calls: %s", count, got)
			}
			if count := strings.Count(got, "state="+test.wantState); count != 2 {
				t.Fatalf("%s status count = %d, want 2; calls: %s", test.wantState, count, got)
			}
			if strings.Count(got, "context=release-intent") != 4 {
				t.Fatalf("release-intent status count is wrong; calls: %s", got)
			}
		})
	}
}

func TestReleasePolicyPlansHighestRequiredVersion(t *testing.T) {
	repo := newReleaseRepository(t)
	output, err := runReleasePolicyPlan(t, repo, repo.head, map[string]string{
		"PR_10_LABEL": "release:minor",
		"PR_11_LABEL": "release:patch",
	})
	if err != nil {
		t.Fatalf("plan-auto error = %v", err)
	}
	want := map[string]string{
		"release":       "true",
		"tag":           "v1.40.0",
		"previous_tag":  "v1.39.6",
		"level":         "minor",
		"pull_requests": "10,11",
		"existing_tag":  "false",
		"release_sha":   repo.head,
	}
	assertReleaseOutput(t, output, want)
}

func TestReleasePolicySkipsChangesMarkedNone(t *testing.T) {
	repo := newReleaseRepository(t)
	output, err := runReleasePolicyPlan(t, repo, repo.head, map[string]string{
		"PR_10_LABEL": "release:none",
		"PR_11_LABEL": "release:none",
	})
	if err != nil {
		t.Fatalf("plan-auto error = %v", err)
	}
	assertReleaseOutput(t, output, map[string]string{
		"release": "false",
		"reason":  "no-release-intent",
	})
}

func TestReleasePolicyFailsClosedForMissingIntent(t *testing.T) {
	repo := newReleaseRepository(t)
	output, err := runReleasePolicyPlan(t, repo, repo.head, map[string]string{
		"PR_11_LABEL": "release:patch",
	})
	if err == nil || !strings.Contains(output, "pull request #10 must have exactly one release label") {
		t.Fatalf("plan-auto error = %v, output = %q", err, output)
	}
}

func TestReleasePolicyFailsClosedForDirectCommit(t *testing.T) {
	repo := newReleaseRepository(t)
	writeReleaseFile(t, repo.dir, "direct", "direct")
	runGit(t, repo.dir, "add", "fixture")
	runGit(t, repo.dir, "commit", "-m", "direct change without pull request")
	directSHA := strings.TrimSpace(runGit(t, repo.dir, "rev-parse", "HEAD"))
	runGit(t, repo.dir, "update-ref", "refs/remotes/origin/master", directSHA)

	output, err := runReleasePolicyPlan(t, repo, directSHA, map[string]string{
		"PR_10_LABEL": "release:minor",
		"PR_11_LABEL": "release:patch",
	})
	if err == nil || !strings.Contains(output, "must identify exactly one pull request merged into master") {
		t.Fatalf("plan-auto error = %v, output = %q", err, output)
	}
}

func TestReleasePolicyHandlesCoveredAndInterruptedRuns(t *testing.T) {
	for _, test := range []struct {
		name          string
		target        string
		releaseExists bool
		releaseDraft  bool
		want          map[string]string
	}{
		{
			name:   "newer unpublished tag is recovered",
			target: "first",
			want: map[string]string{
				"release":      "true",
				"tag":          "v1.40.0",
				"existing_tag": "true",
				"release_sha":  "head",
			},
		},
		{
			name:   "tag exists without release",
			target: "head",
			want: map[string]string{
				"release":      "true",
				"tag":          "v1.40.0",
				"existing_tag": "true",
				"level":        "resume",
			},
		},
		{
			name:         "draft release is resumed",
			target:       "head",
			releaseDraft: true,
			want: map[string]string{
				"release":      "true",
				"tag":          "v1.40.0",
				"existing_tag": "true",
				"level":        "resume",
			},
		},
		{
			name:          "release already exists",
			target:        "head",
			releaseExists: true,
			want: map[string]string{
				"release": "false",
				"reason":  "already-released",
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newReleaseRepository(t)
			runGit(t, repo.dir, "tag", "-a", "v1.40.0", "-m", "release v1.40.0", repo.head)
			target := repo.head
			if test.target == "first" {
				target = repo.first
			}
			variables := map[string]string{
				"PR_10_LABEL": "release:minor",
				"PR_11_LABEL": "release:patch",
			}
			if test.releaseExists {
				variables["RELEASE_EXISTS"] = "true"
			}
			if test.releaseDraft {
				variables["RELEASE_DRAFT"] = "true"
			}
			output, err := runReleasePolicyPlan(t, repo, target, variables)
			if err != nil {
				t.Fatalf("plan-auto error = %v, output = %q", err, output)
			}
			if test.want["release_sha"] == "head" {
				test.want["release_sha"] = repo.head
			}
			assertReleaseOutput(t, output, test.want)
		})
	}
}

func TestReleasePolicyValidatesPushedTagAgainstLabels(t *testing.T) {
	for _, test := range []struct {
		name        string
		tag         string
		wantFailure bool
		wantMessage string
	}{
		{name: "expected tag", tag: "v1.40.0"},
		{name: "unexpected tag", tag: "v1.99.0", wantFailure: true, wantMessage: "does not match expected tag v1.40.0"},
	} {
		t.Run(test.name, func(t *testing.T) {
			repo := newReleaseRepository(t)
			runGit(t, repo.dir, "tag", "-a", test.tag, "-m", "release "+test.tag, repo.head)
			output, err := runReleasePolicy(t, repo, "plan-tag", map[string]string{
				"PR_10_LABEL": "release:minor",
				"PR_11_LABEL": "release:patch",
				"RELEASE_TAG": test.tag,
			})
			if (err != nil) != test.wantFailure {
				t.Fatalf("plan-tag error = %v, wantFailure = %t; output: %s", err, test.wantFailure, output)
			}
			if test.wantMessage != "" && !strings.Contains(output, test.wantMessage) {
				t.Fatalf("plan-tag output = %q, want %q", output, test.wantMessage)
			}
			if !test.wantFailure {
				assertReleaseOutput(t, output, map[string]string{
					"release":      "true",
					"tag":          "v1.40.0",
					"existing_tag": "true",
				})
			}
		})
	}
}

func TestReleasePolicyFailsClosedForAmbiguousTagState(t *testing.T) {
	repo := newReleaseRepository(t)
	runGit(t, repo.dir, "tag", "-a", "v1.40.0", "-m", "release v1.40.0", repo.head)
	runGit(t, repo.dir, "tag", "-a", "v1.41.0", "-m", "release v1.41.0", repo.head)
	output, err := runReleasePolicyPlan(t, repo, repo.head, map[string]string{
		"PR_10_LABEL": "release:minor",
		"PR_11_LABEL": "release:patch",
	})
	if err == nil || !strings.Contains(output, "multiple unpublished stable tags") {
		t.Fatalf("plan-auto error = %v, output = %q", err, output)
	}
}

func TestReleasePolicyFailsClosedWhenReleaseStateIsUnavailable(t *testing.T) {
	repo := newReleaseRepository(t)
	output, err := runReleasePolicyPlan(t, repo, repo.head, map[string]string{
		"PR_10_LABEL":       "release:minor",
		"PR_11_LABEL":       "release:patch",
		"RELEASE_API_ERROR": "true",
	})
	if err == nil || !strings.Contains(output, "cannot determine whether GitHub Release") {
		t.Fatalf("plan-auto error = %v, output = %q", err, output)
	}
}

type releaseRepository struct {
	dir   string
	first string
	head  string
}

func newReleaseRepository(t *testing.T) releaseRepository {
	t.Helper()
	dir := t.TempDir()
	runGit(t, dir, "init", "-b", "master")
	runGit(t, dir, "config", "user.name", "Release Test")
	runGit(t, dir, "config", "user.email", "release@example.com")
	writeReleaseFile(t, dir, "base", "base")
	runGit(t, dir, "add", "fixture")
	runGit(t, dir, "commit", "-m", "base")
	runGit(t, dir, "tag", "-a", "v1.39.6", "-m", "release v1.39.6")

	writeReleaseFile(t, dir, "feature", "feature")
	runGit(t, dir, "add", "fixture")
	runGit(t, dir, "commit", "-m", "feature commit without pull request suffix")
	first := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))

	writeReleaseFile(t, dir, "fix", "fix")
	runGit(t, dir, "add", "fixture")
	runGit(t, dir, "commit", "-m", "fix commit without pull request suffix")
	head := strings.TrimSpace(runGit(t, dir, "rev-parse", "HEAD"))
	runGit(t, dir, "update-ref", "refs/remotes/origin/master", head)
	return releaseRepository{dir: dir, first: first, head: head}
}

func runReleasePolicyPlan(t *testing.T, repo releaseRepository, target string, variables map[string]string) (string, error) {
	t.Helper()
	variables["RELEASE_TARGET_SHA"] = target
	return runReleasePolicy(t, repo, "plan-auto", variables)
}

func runReleasePolicy(t *testing.T, repo releaseRepository, mode string, variables map[string]string) (string, error) {
	t.Helper()
	script := releasePolicyScript(t)
	writeReleaseFakeGH(t, repo.dir)
	outputPath := filepath.Join(repo.dir, "github-output")
	command := exec.Command("bash", script, mode)
	command.Dir = repo.dir
	command.Env = append(os.Environ(),
		"PATH="+repo.dir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"GITHUB_OUTPUT="+outputPath,
		"GITHUB_REPOSITORY=ucloud/terraform-provider-ucloud",
		"PR_10_SHA="+repo.first,
		"PR_11_SHA="+repo.head,
	)
	for name, value := range variables {
		command.Env = append(command.Env, name+"="+value)
	}
	combined, err := command.CombinedOutput()
	outputs, readErr := os.ReadFile(outputPath)
	if readErr != nil && !os.IsNotExist(readErr) {
		t.Fatalf("read GitHub output: %v", readErr)
	}
	return string(combined) + string(outputs), err
}

func assertReleaseOutput(t *testing.T, output string, want map[string]string) {
	t.Helper()
	got := make(map[string]string)
	for _, line := range strings.Split(output, "\n") {
		if name, value, found := strings.Cut(line, "="); found {
			got[name] = value
		}
	}
	for name, value := range want {
		if got[name] != value {
			t.Errorf("%s = %q, want %q; output: %s", name, got[name], value, output)
		}
	}
}

func releasePolicyScript(t *testing.T) string {
	t.Helper()
	script, err := filepath.Abs("release-policy.sh")
	if err != nil {
		t.Fatal(err)
	}
	return script
}

func writeReleaseEvent(t *testing.T, filename string, labels []string) {
	t.Helper()
	payloadLabels := make([]map[string]string, 0, len(labels))
	for _, label := range labels {
		payloadLabels = append(payloadLabels, map[string]string{"name": label})
	}
	payload := map[string]interface{}{
		"repository": map[string]string{"full_name": "ucloud/terraform-provider-ucloud"},
		"pull_request": map[string]interface{}{
			"head":             map[string]string{"sha": releaseHeadSHA},
			"merge_commit_sha": releaseMergeSHA,
			"labels":           payloadLabels,
		},
	}
	contents, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filename, contents, 0644); err != nil {
		t.Fatal(err)
	}
}

func writeReleaseFakeGH(t *testing.T, dir string) {
	t.Helper()
	contents := `#!/usr/bin/env bash
set -euo pipefail
if [[ "${1:-}" != 'api' ]]; then
  echo "unexpected gh command: $*" >&2
  exit 2
fi
shift
if [[ "${1:-}" == '--method' ]]; then
  printf '%s\n' "api $*" >>"${RELEASE_GH_CALLS}"
  printf '{"ok":true}\n'
  exit 0
fi
endpoint="${1:-}"
if [[ "${endpoint}" == */commits/"${PR_10_SHA:-missing}"/pulls ]]; then
  printf '[{"number":10,"merged_at":"2026-09-08T00:00:00Z","base":{"ref":"master"}}]\n'
  exit 0
fi
if [[ "${endpoint}" == */commits/"${PR_11_SHA:-missing}"/pulls ]]; then
  printf '[{"number":11,"merged_at":"2026-09-08T00:00:00Z","base":{"ref":"master"}}]\n'
  exit 0
fi
case "${endpoint}" in
  */pulls/10) label="${PR_10_LABEL:-}" ;;
  */pulls/11) label="${PR_11_LABEL:-}" ;;
	*/releases/tags/*)
    tag="${endpoint##*/}"
    if [[ "${RELEASE_API_ERROR:-false}" == 'true' ]]; then
      echo 'gh: HTTP 503: service unavailable' >&2
      exit 1
    fi
    if [[ "${tag}" == "${PUBLISHED_TAG:-v1.39.6}" || ("${RELEASE_EXISTS:-false}" == 'true' && "${tag}" == 'v1.40.0') ]]; then
      printf '{"tag_name":"%s","draft":false}\n' "${tag}"
      exit 0
    fi
    if [[ "${RELEASE_DRAFT:-false}" == 'true' && "${tag}" == 'v1.40.0' ]]; then
      printf '{"tag_name":"%s","draft":true}\n' "${tag}"
      exit 0
    fi
    echo 'gh: Not Found (HTTP 404)' >&2
    exit 1
    ;;
  */commits/*/pulls) printf '[]\n'; exit 0 ;;
  *) echo "unexpected gh api endpoint: ${endpoint}" >&2; exit 2 ;;
esac
if [[ -n "${label}" ]]; then
  printf '{"merged_at":"2026-09-08T00:00:00Z","base":{"ref":"master"},"labels":[{"name":"%s"}]}\n' "${label}"
else
  printf '{"merged_at":"2026-09-08T00:00:00Z","base":{"ref":"master"},"labels":[]}\n'
fi
`
	filename := filepath.Join(dir, "gh")
	if err := os.WriteFile(filename, []byte(contents), 0755); err != nil {
		t.Fatal(err)
	}
}

func writeReleaseFile(t *testing.T, dir, value, message string) {
	t.Helper()
	filename := filepath.Join(dir, "fixture")
	if err := os.WriteFile(filename, []byte(value+"\n"), 0644); err != nil {
		t.Fatalf("write %s: %v", message, err)
	}
}

func runGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	command := exec.Command("git", args...)
	command.Dir = dir
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, output)
	}
	return fmt.Sprintf("%s", output)
}
