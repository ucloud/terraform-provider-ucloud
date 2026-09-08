#!/usr/bin/env bash

set -euo pipefail

readonly RELEASE_LABEL_PREFIX='release:'
readonly RELEASE_STATUS_CONTEXT='release-intent'
readonly STABLE_TAG_PATTERN='^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)$'

RELEASE_COMMITS_FILE=''
RELEASE_PULL_REQUESTS_FILE=''

cleanup() {
  if [[ -n "${RELEASE_COMMITS_FILE}" ]]; then
    rm -f -- "${RELEASE_COMMITS_FILE}"
  fi
  if [[ -n "${RELEASE_PULL_REQUESTS_FILE}" ]]; then
    rm -f -- "${RELEASE_PULL_REQUESTS_FILE}"
  fi
}

trap cleanup EXIT

die() {
  printf 'release policy: %s\n' "$*" >&2
  exit 1
}

require_command() {
  command -v "$1" >/dev/null 2>&1 || die "required command $1 is unavailable"
}

require_repository() {
  local repository="$1"
  [[ "${repository}" =~ ^[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+$ ]] ||
    die "GitHub repository ${repository} must be owner/name"
}

require_commit_sha() {
  local sha="$1"
  [[ "${sha}" =~ ^[0-9a-fA-F]{40}$ ]] || die "commit SHA ${sha} is invalid"
}

emit_output() {
  local name="$1"
  local value="$2"
  [[ -n "${GITHUB_OUTPUT:-}" ]] || die 'GITHUB_OUTPUT is empty'
  printf '%s=%s\n' "${name}" "${value}" >>"${GITHUB_OUTPUT}"
}

set_release_status() {
  local repository="$1"
  local sha="$2"
  local state="$3"
  local description="$4"
  local target_url="${GITHUB_SERVER_URL:-https://github.com}/${repository}/actions/runs/${GITHUB_RUN_ID:-0}"

  gh api --method POST "repos/${repository}/statuses/${sha}" \
    -f "state=${state}" \
    -f "context=${RELEASE_STATUS_CONTEXT}" \
    -f "description=${description}" \
    -f "target_url=${target_url}" >/dev/null
}

set_pull_request_statuses() {
  local repository="$1"
  local head_sha="$2"
  local merge_sha="$3"
  local state="$4"
  local description="$5"

  set_release_status "${repository}" "${head_sha}" "${state}" "${description}"
  if [[ -n "${merge_sha}" && "${merge_sha}" != "${head_sha}" ]]; then
    set_release_status "${repository}" "${merge_sha}" "${state}" "${description}"
  fi
}

fail_pull_request_check() {
  local repository="$1"
  local head_sha="$2"
  local merge_sha="$3"
  local message="$4"

  if ! set_pull_request_statuses "${repository}" "${head_sha}" "${merge_sha}" failure 'select exactly one supported release label'; then
    printf 'release policy: failed to publish release-intent failure status\n' >&2
  fi
  die "${message}"
}

check_pull_request() {
  require_command gh
  require_command jq
  [[ -n "${GITHUB_EVENT_PATH:-}" ]] || die 'GITHUB_EVENT_PATH is empty'

  local repository
  local head_sha
  local merge_sha
  repository="$(jq -er '.repository.full_name | strings | select(length > 0)' "${GITHUB_EVENT_PATH}")" ||
    die 'pull request event has no repository.full_name'
  head_sha="$(jq -er '.pull_request.head.sha | strings | select(length > 0)' "${GITHUB_EVENT_PATH}")" ||
    die 'pull request event has no head SHA'
  merge_sha="$(jq -r '.pull_request.merge_commit_sha // ""' "${GITHUB_EVENT_PATH}")"
  require_repository "${repository}"
  require_commit_sha "${head_sha}"
  if [[ -n "${merge_sha}" ]]; then
    require_commit_sha "${merge_sha}"
  fi

  set_pull_request_statuses "${repository}" "${head_sha}" "${merge_sha}" pending 'checking pull request release intent'

  local release_labels
  local label_count
  local release_label
  release_labels="$(jq -c '[.pull_request.labels[]?.name | strings | select(startswith("release:"))]' "${GITHUB_EVENT_PATH}")" ||
    fail_pull_request_check "${repository}" "${head_sha}" "${merge_sha}" 'cannot read pull request labels'
  label_count="$(jq -r 'length' <<<"${release_labels}")"
  if [[ "${label_count}" != '1' ]]; then
    fail_pull_request_check "${repository}" "${head_sha}" "${merge_sha}" \
      "pull request must have exactly one release label; found ${label_count}"
  fi

  release_label="$(jq -r '.[0]' <<<"${release_labels}")"
  case "${release_label}" in
    release:major | release:minor | release:patch | release:none) ;;
    *)
      fail_pull_request_check "${repository}" "${head_sha}" "${merge_sha}" \
        "unsupported release label ${release_label}"
      ;;
  esac

  set_pull_request_statuses "${repository}" "${head_sha}" "${merge_sha}" success \
    "release intent is ${release_label#"${RELEASE_LABEL_PREFIX}"}"
  printf 'release policy: %s\n' "${release_label}"
}

stable_tags_merged_into() {
  local ref="$1"
  local tag
  while IFS= read -r tag; do
    if [[ "${tag}" =~ ${STABLE_TAG_PATTERN} ]]; then
      printf '%s\n' "${tag}"
    fi
  done < <(git tag --merged "${ref}" --sort=-version:refname)
}

latest_stable_tag() {
  local ref="$1"
  local tag
  while IFS= read -r tag; do
    printf '%s\n' "${tag}"
    return 0
  done < <(stable_tags_merged_into "${ref}")
  return 1
}

latest_published_tag() {
  local repository="$1"
  local ref="$2"
  local tag
  while IFS= read -r tag; do
    release_exists "${repository}" "${tag}"
    if [[ "${RELEASE_PUBLISHED}" == 'true' ]]; then
      printf '%s\n' "${tag}"
      return 0
    fi
  done < <(stable_tags_merged_into "${ref}")
  return 1
}

unpublished_tags_after() {
  local ref="$1"
  local published_tag="$2"
  local tag
  while IFS= read -r tag; do
    if [[ "${tag}" == "${published_tag}" ]]; then
      return 0
    fi
    printf '%s\n' "${tag}"
  done < <(stable_tags_merged_into "${ref}")
  die "published tag ${published_tag} is not merged into ${ref}"
}

release_exists() {
  local repository="$1"
  local tag="$2"
  local response

  if response="$(gh api "repos/${repository}/releases/tags/${tag}" 2>&1)"; then
    RELEASE_DRAFT="$(jq -r 'if (.draft | type) == "boolean" then .draft else error("draft must be boolean") end' <<<"${response}")" ||
      die "GitHub Release ${tag} has no valid draft state"
    if [[ "${RELEASE_DRAFT}" == 'false' ]]; then
      RELEASE_PUBLISHED='true'
    else
      RELEASE_PUBLISHED='false'
    fi
  elif [[ "${response}" == *'HTTP 404'* ]]; then
    RELEASE_DRAFT='false'
    RELEASE_PUBLISHED='false'
  else
    die "cannot determine whether GitHub Release ${tag} exists: ${response}"
  fi
}

pull_request_number_for_commit() {
  local repository="$1"
  local commit="$2"
  local associations
  local pull_requests
  local association_count

  associations="$(gh api "repos/${repository}/commits/${commit}/pulls")" ||
    die "cannot read pull request associations for commit ${commit}"
  pull_requests="$(jq -c '[.[] | select(.merged_at != null and .base.ref == "master") | .number] | unique' <<<"${associations}")" ||
    die "cannot parse pull request associations for commit ${commit}"
  association_count="$(jq -r 'length' <<<"${pull_requests}")"
  if [[ "${association_count}" != '1' ]]; then
    die "commit ${commit} must identify exactly one pull request merged into master; found ${association_count}"
  fi
  jq -r '.[0]' <<<"${pull_requests}"
}

release_level_for_pull_request() {
  local repository="$1"
  local pull_request="$2"
  local metadata
  local label_count
  local release_label

  metadata="$(gh api "repos/${repository}/pulls/${pull_request}")" ||
    die "cannot read pull request #${pull_request}"
  if [[ "$(jq -r '.merged_at // ""' <<<"${metadata}")" == '' ||
    "$(jq -r '.base.ref // ""' <<<"${metadata}")" != 'master' ]]; then
    die "pull request #${pull_request} is not merged into master"
  fi

  label_count="$(jq '[.labels[]?.name | strings | select(startswith("release:"))] | length' <<<"${metadata}")" ||
    die "cannot read labels for pull request #${pull_request}"
  if [[ "${label_count}" != '1' ]]; then
    die "pull request #${pull_request} must have exactly one release label; found ${label_count}"
  fi
  release_label="$(jq -r '[.labels[]?.name | strings | select(startswith("release:"))][0]' <<<"${metadata}")"
  case "${release_label}" in
    release:major) printf 'major\n' ;;
    release:minor) printf 'minor\n' ;;
    release:patch) printf 'patch\n' ;;
    release:none) printf 'none\n' ;;
    *) die "pull request #${pull_request} has unsupported release label ${release_label}" ;;
  esac
}

next_tag() {
  local previous_tag="$1"
  local level="$2"
  [[ "${previous_tag}" =~ ${STABLE_TAG_PATTERN} ]] || die "tag ${previous_tag} is not stable SemVer"
  local major="${BASH_REMATCH[1]}"
  local minor="${BASH_REMATCH[2]}"
  local patch="${BASH_REMATCH[3]}"

  case "${level}" in
    major)
      major=$((major + 1))
      minor=0
      patch=0
      ;;
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    patch) patch=$((patch + 1)) ;;
    *) die "cannot calculate a tag for release level ${level}" ;;
  esac
  printf 'v%s.%s.%s\n' "${major}" "${minor}" "${patch}"
}

evaluate_release_range() {
  local repository="$1"
  local previous_tag="$2"
  local target_sha="$3"
  local commit
  local pull_request
  local level
  local level_rank
  local highest_level='none'
  local highest_rank=0

  RELEASE_COMMITS_FILE="$(mktemp)"
  RELEASE_PULL_REQUESTS_FILE="$(mktemp)"
  git rev-list --first-parent --reverse "${previous_tag}..${target_sha}" >"${RELEASE_COMMITS_FILE}"
  [[ -s "${RELEASE_COMMITS_FILE}" ]] || die "tag ${previous_tag} already points at release target ${target_sha}"

  while IFS= read -r commit; do
    pull_request="$(pull_request_number_for_commit "${repository}" "${commit}")"
    printf '%s\n' "${pull_request}" >>"${RELEASE_PULL_REQUESTS_FILE}"
  done <"${RELEASE_COMMITS_FILE}"
  sort -nu -o "${RELEASE_PULL_REQUESTS_FILE}" "${RELEASE_PULL_REQUESTS_FILE}"

  RELEASE_PULL_REQUESTS=''
  while IFS= read -r pull_request; do
    level="$(release_level_for_pull_request "${repository}" "${pull_request}")"
    case "${level}" in
      major) level_rank=3 ;;
      minor) level_rank=2 ;;
      patch) level_rank=1 ;;
      none) level_rank=0 ;;
      *) die "pull request #${pull_request} returned invalid release level ${level}" ;;
    esac
    if ((level_rank > highest_rank)); then
      highest_rank="${level_rank}"
      highest_level="${level}"
    fi
    if [[ -z "${RELEASE_PULL_REQUESTS}" ]]; then
      RELEASE_PULL_REQUESTS="${pull_request}"
    else
      RELEASE_PULL_REQUESTS="${RELEASE_PULL_REQUESTS},${pull_request}"
    fi
  done <"${RELEASE_PULL_REQUESTS_FILE}"

  RELEASE_LEVEL="${highest_level}"
  if [[ "${highest_level}" == 'none' ]]; then
    RELEASE_EXPECTED_TAG=''
  else
    RELEASE_EXPECTED_TAG="$(next_tag "${previous_tag}" "${highest_level}")"
  fi
}

validate_plan_environment() {
  require_command gh
  require_command git
  require_command jq
  [[ -n "${GITHUB_REPOSITORY:-}" ]] || die 'GITHUB_REPOSITORY is empty'
  require_repository "${GITHUB_REPOSITORY}"
  git rev-parse --verify 'refs/remotes/origin/master^{commit}' >/dev/null 2>&1 ||
    die 'refs/remotes/origin/master is unavailable'
}

validate_release_target() {
  local target_sha="$1"
  require_commit_sha "${target_sha}"
  git cat-file -e "${target_sha}^{commit}" 2>/dev/null || die "release target ${target_sha} is unavailable"
  git merge-base --is-ancestor "${target_sha}" refs/remotes/origin/master ||
    die "release target ${target_sha} is not on master"
}

emit_skip() {
  emit_output release false
  emit_output reason "$1"
}

plan_existing_tag() {
  local repository="$1"
  local tag="$2"
  local target_sha="$3"
  local previous_tag="$4"

  release_exists "${repository}" "${tag}"
  if [[ "${RELEASE_PUBLISHED}" == 'true' ]]; then
    emit_skip already-released
    return 0
  fi

  [[ "${previous_tag}" =~ ${STABLE_TAG_PATTERN} ]] || die "previous tag ${previous_tag} is not stable SemVer"
  local previous_tag_sha
  previous_tag_sha="$(git rev-list -n 1 "${previous_tag}")"
  git merge-base --is-ancestor "${previous_tag_sha}" "${target_sha}" ||
    die "tag ${tag} is not descended from published tag ${previous_tag}"
  evaluate_release_range "${repository}" "${previous_tag}" "${target_sha}"
  if [[ "${RELEASE_LEVEL}" == 'none' ]]; then
    die "tag ${tag} exists for changes that are all marked release:none"
  fi
  if [[ "${RELEASE_EXPECTED_TAG}" != "${tag}" ]]; then
    die "tag ${tag} does not match expected tag ${RELEASE_EXPECTED_TAG}"
  fi

  emit_output release true
  emit_output tag "${tag}"
  emit_output previous_tag "${previous_tag}"
  emit_output level resume
  emit_output pull_requests "${RELEASE_PULL_REQUESTS}"
  emit_output existing_tag true
  emit_output release_sha "${target_sha}"
}

plan_automatic_release() {
  validate_plan_environment
  [[ -n "${RELEASE_TARGET_SHA:-}" ]] || die 'RELEASE_TARGET_SHA is empty'
  validate_release_target "${RELEASE_TARGET_SHA}"

  local latest_tag
  local published_tag
  local published_tag_sha
  local unpublished_tags
  local unpublished_count
  local pending_tag
  local pending_tag_sha
  latest_tag="$(latest_stable_tag refs/remotes/origin/master)" ||
    die 'no stable SemVer tag is merged into master'
  published_tag="$(latest_published_tag "${GITHUB_REPOSITORY}" refs/remotes/origin/master)" ||
    die 'no published stable SemVer tag is merged into master'
  published_tag_sha="$(git rev-list -n 1 "${published_tag}")"

  unpublished_tags="$(unpublished_tags_after refs/remotes/origin/master "${published_tag}" | jq -Rsc 'split("\n") | map(select(length > 0))')"
  unpublished_count="$(jq -r 'length' <<<"${unpublished_tags}")"
  if ((unpublished_count > 1)); then
    die "multiple unpublished stable tags exist after ${published_tag}: $(jq -r 'join(",")' <<<"${unpublished_tags}")"
  fi
  if [[ "${unpublished_count}" == '1' ]]; then
    pending_tag="$(jq -r '.[0]' <<<"${unpublished_tags}")"
    pending_tag_sha="$(git rev-list -n 1 "${pending_tag}")"
    plan_existing_tag "${GITHUB_REPOSITORY}" "${pending_tag}" "${pending_tag_sha}" "${published_tag}"
    return 0
  fi

  [[ "${latest_tag}" == "${published_tag}" ]] || die "release tag state changed while planning"
  if git merge-base --is-ancestor "${RELEASE_TARGET_SHA}" "${published_tag_sha}"; then
    if [[ "${RELEASE_TARGET_SHA}" == "${published_tag_sha}" ]]; then
      emit_skip already-released
    else
      emit_skip "covered-by-${published_tag}"
    fi
    return 0
  fi
  git merge-base --is-ancestor "${published_tag_sha}" "${RELEASE_TARGET_SHA}" ||
    die "release target ${RELEASE_TARGET_SHA} and ${published_tag} have diverged"

  evaluate_release_range "${GITHUB_REPOSITORY}" "${published_tag}" "${RELEASE_TARGET_SHA}"
  if [[ "${RELEASE_LEVEL}" == 'none' ]]; then
    emit_skip no-release-intent
    return 0
  fi

  emit_output release true
  emit_output tag "${RELEASE_EXPECTED_TAG}"
  emit_output previous_tag "${published_tag}"
  emit_output level "${RELEASE_LEVEL}"
  emit_output pull_requests "${RELEASE_PULL_REQUESTS}"
  emit_output existing_tag false
  emit_output release_sha "${RELEASE_TARGET_SHA}"
}

plan_tag_release() {
  validate_plan_environment
  [[ -n "${RELEASE_TAG:-}" ]] || die 'RELEASE_TAG is empty'
  [[ "${RELEASE_TAG}" =~ ${STABLE_TAG_PATTERN} ]] || die "tag ${RELEASE_TAG} is not stable SemVer"
  git rev-parse --verify "refs/tags/${RELEASE_TAG}^{commit}" >/dev/null 2>&1 ||
    die "tag ${RELEASE_TAG} is unavailable"
  local target_sha
  target_sha="$(git rev-list -n 1 "refs/tags/${RELEASE_TAG}")"
  validate_release_target "${target_sha}"

  release_exists "${GITHUB_REPOSITORY}" "${RELEASE_TAG}"
  if [[ "${RELEASE_PUBLISHED}" == 'true' ]]; then
    emit_skip already-released
    return 0
  fi
  local published_tag
  local unpublished_tags
  local unpublished_count
  published_tag="$(latest_published_tag "${GITHUB_REPOSITORY}" refs/remotes/origin/master)" ||
    die 'no published stable SemVer tag is merged into master'
  unpublished_tags="$(unpublished_tags_after refs/remotes/origin/master "${published_tag}" | jq -Rsc 'split("\n") | map(select(length > 0))')"
  unpublished_count="$(jq -r 'length' <<<"${unpublished_tags}")"
  if [[ "${unpublished_count}" != '1' || "$(jq -r '.[0] // ""' <<<"${unpublished_tags}")" != "${RELEASE_TAG}" ]]; then
    die "tag ${RELEASE_TAG} must be the only unpublished stable tag after ${published_tag}"
  fi
  plan_existing_tag "${GITHUB_REPOSITORY}" "${RELEASE_TAG}" "${target_sha}" "${published_tag}"
}

case "${1:-}" in
  check-pr) check_pull_request ;;
  plan-auto) plan_automatic_release ;;
  plan-tag) plan_tag_release ;;
  *) die 'usage: release-policy.sh {check-pr|plan-auto|plan-tag}' ;;
esac
