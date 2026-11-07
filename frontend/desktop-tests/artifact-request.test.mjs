import test from 'node:test';
import assert from 'node:assert/strict';
import {spawnSync} from 'node:child_process';
import {fileURLToPath} from 'node:url';

const guard = fileURLToPath(new URL('../../scripts/validate-desktop-artifact-request.sh', import.meta.url));
const source = 'a'.repeat(40);

function request(overrides = {}) {
  return spawnSync('sh', [guard], {encoding: 'utf8', timeout: 5000, env: {
    ...process.env, GITHUB_EVENT_NAME: 'workflow_dispatch', GITHUB_REF: 'refs/heads/main',
    GITHUB_RUN_ATTEMPT: '1', GITHUB_SHA: source, CHARTROVE_ARTIFACT_SOURCE_SHA: source, ...overrides,
  }});
}

test('accepts a first manual main request bound to the exact selected revision', () => {
  const result = request();
  assert.equal(result.status, 0, result.stderr);
  assert.equal(result.stdout, '');
});

for (const [name, overrides] of [
  ['automatic push', {GITHUB_EVENT_NAME: 'push'}],
  ['pull request', {GITHUB_EVENT_NAME: 'pull_request', GITHUB_REF: 'refs/pull/1/merge'}],
  ['another branch', {GITHUB_REF: 'refs/heads/task/synthetic'}],
  ['a different selected revision', {GITHUB_SHA: 'b'.repeat(40)}],
  ['an unbound revision', {CHARTROVE_ARTIFACT_SOURCE_SHA: ''}],
  ['a malformed revision', {CHARTROVE_ARTIFACT_SOURCE_SHA: 'not-a-revision'}],
  ['a reused workflow attempt', {GITHUB_RUN_ATTEMPT: '2'}],
]) {
  test(`rejects ${name} before an artifact can be prepared`, () => {
    const result = request(overrides);
    assert.equal(result.status, 1, result.stderr);
    assert.equal(result.stdout, '');
  });
}
