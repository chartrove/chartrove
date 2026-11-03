import assert from 'node:assert/strict';
import { existsSync, mkdirSync, mkdtempSync, readFileSync, rmSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import test from 'node:test';
import { validate } from '../tools/validate.mjs';

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
