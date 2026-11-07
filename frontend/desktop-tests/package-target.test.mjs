import test from 'node:test';
import assert from 'node:assert/strict';
import { spawnSync } from 'node:child_process';
import { copyFileSync, existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';

const buildScript = fileURLToPath(new URL('../../desktop/build.sh', import.meta.url));

function probe({ system, machine, ...target }) {
  const root = mkdtempSync(join(tmpdir(), 'chartrove-package target-'));
  try {
    const desktop = join(root, 'desktop');
    const tools = join(root, 'tools');
    mkdirSync(desktop);
    mkdirSync(tools);
    mkdirSync(join(root, 'frontend/node_modules/.bin'), { recursive: true });
    copyFileSync(buildScript, join(desktop, 'build.sh'));
    writeFileSync(join(root, 'frontend/node_modules/.bin/ng'), '#!/bin/sh\nexit 99\n', { mode: 0o700 });
    writeFileSync(join(tools, 'uname'), '#!/bin/sh\ncase "$1" in -s) printf "%s\\n" "$FAKE_SYSTEM" ;; -m) printf "%s\\n" "$FAKE_MACHINE" ;; *) exit 99 ;; esac\n', { mode: 0o700 });
    // Stop at the first frontend producer; no actual build or SDK is invoked.
    writeFileSync(join(tools, 'node'), '#!/bin/sh\nprintf "frontend producer\\n" > "$PRODUCER_LOG"\nexit 91\n', { mode: 0o700 });
    const log = join(root, 'producer.log');
    const result = spawnSync('sh', [join(desktop, 'build.sh')], {
      cwd: root,
      env: { PATH: `${tools}:${process.env.PATH}`, TMPDIR: root,
        FAKE_SYSTEM: system, FAKE_MACHINE: machine, PRODUCER_LOG: log, ...target },
      encoding: 'utf8', timeout: 10_000,
    });
    assert.ifError(result.error);
    return { status: result.status, stderr: result.stderr,
      producer: existsSync(log) ? readFileSync(log, 'utf8') : '' };
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
}

for (const [name, platform] of [
  ['Linux host', { system: 'Linux', machine: 'x86_64' }],
  ['Intel Mac', { system: 'Darwin', machine: 'x86_64' }],
  ['conflicting Go architecture', { system: 'Darwin', machine: 'arm64', GOARCH: 'amd64' }],
  ['conflicting Go OS', { system: 'Darwin', machine: 'arm64', GOOS: 'linux' }],
]) {
  test(`desktop packaging rejects ${name} before any build`, () => {
    const result = probe(platform);
    assert.equal(result.status, 2);
    assert.match(result.stderr, /macOS Apple Silicon/);
    assert.equal(result.producer, '');
  });
}

test('native Apple Silicon target reaches the existing build boundary', () => {
  const result = probe({ system: 'Darwin', machine: 'arm64', GOOS: 'darwin', GOARCH: 'arm64' });
  assert.equal(result.status, 91);
  assert.equal(result.producer, 'frontend producer\n');
});
