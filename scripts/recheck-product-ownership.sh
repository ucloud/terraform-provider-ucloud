#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY must be set}"
mode="${1:-check}"
if [[ "${mode}" != 'check' && "${mode}" != 'list' ]]; then
  echo 'expected check or list mode' >&2
  exit 2
fi
review_dir="$(mktemp -d "${RUNNER_TEMP:-${TMPDIR:-/tmp}}/product-ownership-review.XXXXXX")"
trap 'rm -rf "${review_dir}"' EXIT

if [[ "${mode}" == 'check' ]]; then
  go build -o "${review_dir}/product-ownership" ./cmd/product-ownership
fi

# workflow_run may omit pull_requests for forks. Read live PRs from GitHub
# instead of trusting artifacts, branch content, or a relayed event payload.
if [[ -n "${PRODUCT_OWNERSHIP_PR:-}" ]]; then
  [[ "${PRODUCT_OWNERSHIP_PR}" =~ ^[1-9][0-9]*$ ]] || exit 2
  numbers="${PRODUCT_OWNERSHIP_PR}"
else
  numbers="$(gh api --paginate "repos/${GITHUB_REPOSITORY}/pulls?state=open&base=master&per_page=100" --jq '.[].number')"
fi
result=0
while IFS= read -r number; do
  [[ -n "${number}" ]] || continue
  gh api "repos/${GITHUB_REPOSITORY}/pulls/${number}" >"${review_dir}/pull-request.json"
  if ! jq -e '.state == "open" and .base.ref == "master" and .changed_files == 1' "${review_dir}/pull-request.json" >/dev/null; then
    continue
  fi
  author="$(jq -r '.user.login' "${review_dir}/pull-request.json")"
  # Core-authored PRs keep the original event-sender check. A review must not
  # turn an unauthorized push to a Core author's branch into an authorized one.
  if jq -e --arg author "${author}" '.core.github_users | map(ascii_downcase) | index($author | ascii_downcase) != null' .github/product-owners.json >/dev/null; then
    continue
  fi
  filename="$(gh api "repos/${GITHUB_REPOSITORY}/pulls/${number}/files" --jq '.[].filename')"
  [[ "${filename}" == '.github/product-owners.json' ]] || continue
  if [[ "${mode}" == 'list' ]]; then
    printf '%s\n' "${number}"
    continue
  fi

  # A review does not change the proposal. Revalidate its current author and
  # contents; the API review records determine the approving Core identity.
  jq --arg repository "${GITHUB_REPOSITORY}" \
    '{repository: {full_name: $repository}, sender: .user, pull_request: .}' \
    "${review_dir}/pull-request.json" >"${review_dir}/event.json"
  if ! "${review_dir}/product-ownership" -config .github/product-owners.json -event "${review_dir}/event.json"; then
    result=1
  fi
done <<<"${numbers}"
exit "${result}"
