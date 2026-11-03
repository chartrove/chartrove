import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import test from 'node:test';
import { validate } from '../tools/validate.mjs';

// Linux CI provides a separate tmpfs. Exercise real EXDEV behavior rather than
// mocking rename; hosts without a second filesystem retain the ordinary cases.
const separateFilesystem = existsSync('/dev/shm') && statSync('/dev/shm').dev !== statSync(tmpdir()).dev;
for (const failure of [false, true]) {
  test(`validation ${failure ? 'failure' : 'success'} restores documentation across filesystems`,
    { skip: !separateFilesystem }, async () => {
      const root = mkdtempSync('/dev/shm/chartrove-retention-test-');
      try {
        mkdirSync(join(root, 'frontend', 'desktop-tests'), { recursive: true });
        const documentation = join(root, 'frontend', 'documentation.json');
        writeFileSync(documentation, 'original across-volume documentation', { mode: 0o640 });
        const before = statSync(documentation);
        let commands = 0;
        const result = await validate({ root, reportDir: join(root, 'report'), run: async (cmd, cwd, env, log) => {
          assert.notEqual(statSync(root).dev, statSync(env.TMPDIR).dev);
          commands++;
          writeFileSync(join(cwd, 'documentation.json'), 'generated documentation');
          writeFileSync(log, 'command reached across filesystem boundary');
          return failure ? 7 : 0;
        } });
        assert.equal(result, failure ? 7 : 0);
        assert.equal(commands, failure ? 1 : 4);
        assert.equal(readFileSync(documentation, 'utf8'), 'original across-volume documentation');
        const after = statSync(documentation);
        assert.equal(after.ino, before.ino);
        assert.equal(after.mode, before.mode);
        assert.equal(existsSync(join(root, 'frontend', '.chartrove-validation.lock')), false);
        assert.equal(JSON.parse(readFileSync(join(root, 'report', 'receipt.json'))).scratchRemoved, true);
      } finally { rmSync(root, { recursive: true, force: true }); }
    });
}

for (const failure of [false, true]) {
  test(`validation ${failure ? 'failure' : 'success'} retains evidence and removes generated payload`, async () => {
    const root = mkdtempSync(join(tmpdir(), 'chartrove-retention-test-'));
    try {
      mkdirSync(join(root, 'frontend', 'desktop-tests'), { recursive: true });
      writeFileSync(join(root, 'frontend', 'documentation.json'), 'original documentation');
      const reportDir = join(root, 'report');
      let scratch;
      const result = await validate({ root, reportDir, run: async (cmd, cwd, env, log) => {
        scratch = env.TMPDIR;
        const output = join(scratch, 'frontend');
        mkdirSync(output, { recursive: true });
        writeFileSync(join(output, 'index.html'), 'disposable app');
        writeFileSync(join(cwd, 'documentation.json'), 'generated documentation');
        writeFileSync(log, 'retained command log');
        return failure ? 7 : 0;
      } });
      assert.equal(result, failure ? 7 : 0);
      assert.equal(existsSync(scratch), false);
      assert.equal(readFileSync(join(root, 'frontend', 'documentation.json'), 'utf8'), 'original documentation');
      const receipt = JSON.parse(readFileSync(join(reportDir, 'receipt.json')));
      assert.equal(receipt.scratchRemoved, true);
      assert.ok(Number.isSafeInteger(receipt.freeBeforeBytes));
      assert.ok(Number.isSafeInteger(receipt.freeAfterCleanupBytes));
      assert.equal(receipt.freeDeltaBytes, receipt.freeAfterCleanupBytes - receipt.freeBeforeBytes);
      for (const step of receipt.steps) {
        assert.ok(Number.isSafeInteger(step.freeBeforeBytes));
        assert.ok(Number.isSafeInteger(step.freeAfterBytes));
        assert.equal(step.freeDeltaBytes, step.freeAfterBytes - step.freeBeforeBytes);
      }
      assert.equal(receipt.steps.length, failure ? 1 : 4);
      assert.equal(receipt.outputs[0].files[0].path, 'index.html');
      assert.equal(readFileSync(join(reportDir, 'node.log'), 'utf8'), 'retained command log');
      await assert.rejects(validate({ root, reportDir }), { code: 'EEXIST' });
    } finally { rmSync(root, { recursive: true, force: true }); }
  });
}


test('an unexpected command error still restores documentation and removes scratch', async () => {
  const root = mkdtempSync(join(tmpdir(), 'chartrove-retention-test-'));
  try {
    mkdirSync(join(root, 'frontend', 'desktop-tests'), { recursive: true });
    let scratch;
    const reportDir = join(root, 'report');
    await assert.rejects(validate({ root, reportDir, run: async (cmd, cwd, env) => {
      scratch = env.TMPDIR;
      writeFileSync(join(cwd, 'documentation.json'), 'generated');
      throw new Error('spawn failed');
    } }), /spawn failed/);
    assert.equal(existsSync(scratch), false);
    assert.equal(existsSync(join(root, 'frontend', 'documentation.json')), false);
    assert.equal(existsSync(join(root, 'frontend', '.chartrove-validation.lock')), false);
    assert.equal(JSON.parse(readFileSync(join(reportDir, 'receipt.json'))).scratchRemoved, true);
  } finally { rmSync(root, { recursive: true, force: true }); }
});

test('another validation lock prevents modifying existing documentation', async () => {
  const root = mkdtempSync(join(tmpdir(), 'chartrove-retention-test-'));
  try {
    mkdirSync(join(root, 'frontend', '.chartrove-validation.lock'), { recursive: true });
    writeFileSync(join(root, 'frontend', 'documentation.json'), 'untouched');
    await assert.rejects(validate({ root, reportDir: join(root, 'report') }), /Another validation owns/);
    assert.equal(readFileSync(join(root, 'frontend', 'documentation.json'), 'utf8'), 'untouched');
  } finally { rmSync(root, { recursive: true, force: true }); }
});
