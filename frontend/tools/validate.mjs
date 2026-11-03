#!/usr/bin/env node
// Keep validation evidence without retaining another expanded app per run.
import { spawn } from 'node:child_process';
import { createHash } from 'node:crypto';
import { closeSync, existsSync, lstatSync, mkdirSync, mkdtempSync, openSync, readFileSync,
  readdirSync, renameSync, rmSync, statfsSync, statSync, writeFileSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, relative, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

function inventory(root) {
  const files = [];
  function visit(dir) {
    for (const entry of readdirSync(dir, { withFileTypes: true })) {
      const path = join(dir, entry.name);
      if (entry.isDirectory()) visit(path);
      else if (entry.isFile()) files.push({ path: relative(root, path), bytes: statSync(path).size });
    }
  }
  visit(root);
  return files;
}

function freeBytes(root) {
  const disk = statfsSync(root);
  return disk.bavail * disk.bsize;
}

export function runCommand(command, cwd, env, log) {
  return new Promise((resolveCommand, reject) => {
    const fd = openSync(log, 'wx');
    const child = spawn(command[0], command.slice(1), {
      cwd, env, stdio: ['ignore', fd, fd], detached: process.platform !== 'win32',
    });
    let timer;
    let interrupted = false;
    function interrupt() {
      interrupted = true;
      if (!child.pid) return;
      const target = process.platform === 'win32' ? child.pid : -child.pid;
      try { process.kill(target, 'SIGTERM'); } catch (error) {
        if (error.code !== 'ESRCH') throw error;
      }
      timer = setTimeout(() => {
        try { process.kill(target, 'SIGKILL'); } catch (error) {
          if (error.code !== 'ESRCH') throw error;
        }
      }, 5000);
    }
    process.once('SIGINT', interrupt);
    process.once('SIGTERM', interrupt);
    function finish() {
      clearTimeout(timer);
      process.removeListener('SIGINT', interrupt);
      process.removeListener('SIGTERM', interrupt);
      closeSync(fd);
    }
    child.once('error', error => { finish(); reject(error); });
    child.once('exit', (code, signal) => {
      const complete = () => {
        finish();
        resolveCommand(interrupted ? 143 : (code ?? (signal === 'SIGINT' ? 130 : 143)));
      };
      if (interrupted && process.platform !== 'win32') {
        // The CLI can exit before a browser helper. Wait for the group before
        // disposing its scratch; terminate remaining helpers after the grace period.
        const deadline = Date.now() + 5000;
        const waitForGroup = () => {
          try { process.kill(-child.pid, 0); } catch (error) {
            if (error.code === 'ESRCH') { complete(); return; }
            throw error;
          }
          if (Date.now() >= deadline) {
            try { process.kill(-child.pid, 'SIGKILL'); } catch (error) {
              if (error.code !== 'ESRCH') throw error;
            }
            setTimeout(complete, 100);
          } else setTimeout(waitForGroup, 100);
        };
        waitForGroup();
      } else complete();
    });
  });
}

export async function validate({ root, reportDir, run = runCommand }) {
  // Refuse existing evidence directories rather than replacing a retained run.
  mkdirSync(reportDir, { recursive: false, mode: 0o700 });
  const lock = join(root, 'frontend', '.chartrove-validation.lock');
  try { mkdirSync(lock, { mode: 0o700 }); } catch (error) {
    if (error.code === 'EEXIST') throw new Error('Another validation owns frontend/.chartrove-validation.lock; verify it has exited before removing its lock.');
    throw error;
  }
  writeFileSync(join(lock, 'owner.json'), JSON.stringify({ pid: process.pid, started: new Date().toISOString() }) + '\n');
  let scratch;
  try { scratch = mkdtempSync(join(tmpdir(), 'chartrove-validation-')); } catch (error) {
    rmSync(lock, { recursive: true });
    throw error;
  }
  const frontend = join(root, 'frontend');
  const documentation = join(frontend, 'documentation.json');
  // The checkout may be on a different volume from OS scratch. Keep the
  // original beside it so both preservation/restoration renames stay local.
  const originalDocumentation = join(lock, 'documentation-before');
  const receipt = { started: new Date().toISOString(), scratch, freeBeforeBytes: freeBytes(root), steps: [], outputs: [] };
  const save = () => writeFileSync(join(reportDir, 'receipt.json'), JSON.stringify(receipt, null, 2) + '\n');
  const hadDocumentation = (() => {
    try { lstatSync(documentation); return true; } catch (error) {
      if (error.code === 'ENOENT') return false;
      throw error;
    }
  })();
  // Preserve any pre-existing documentation byte-for-byte with a rename.
  try {
    if (hadDocumentation) renameSync(documentation, originalDocumentation);
  } catch (error) {
    rmSync(scratch, { recursive: true, force: true });
    rmSync(lock, { recursive: true });
    throw error;
  }
  const env = { ...process.env, TMPDIR: scratch, NG_CLI_ANALYTICS: 'false',
    CHARTROVE_COVERAGE_DIR: join(scratch, 'coverage') };
  const cli = ['node', 'node_modules/@angular/cli/bin/ng.js'];
  const steps = [
    ['node', ['node', '--test', ...readdirSync(join(frontend, 'desktop-tests'))
      .filter(name => name.endsWith('.test.mjs')).map(name => join('desktop-tests', name))]],
    ['browser', [...cli, 'test', '--watch=false', '--progress=false', '--browsers=ChromeHeadlessCI']],
    ['production', [...cli, 'build', '--configuration', 'desktop_prod', '--output-path', join(scratch, 'frontend')]],
    ['storybook', ['npm', 'exec', '--yes', '--package=corepack@0.35.0', '--', 'corepack',
      'yarn', 'exec', 'ng', '--', 'run', 'fastenhealth:build-storybook', '--output-dir', join(scratch, 'storybook')]],
  ];
  let result = 0;
  try {
    save();
    for (const [name, command] of steps) {
      const step = { name, command, started: new Date().toISOString(), freeBeforeBytes: freeBytes(root) };
      receipt.steps.push(step);
      save();
      step.exit = await run(command, frontend, env, join(reportDir, `${name}.log`));
      step.freeAfterBytes = freeBytes(root);
      step.freeDeltaBytes = step.freeAfterBytes - step.freeBeforeBytes;
      step.finished = new Date().toISOString();
      save();
      if (step.exit !== 0) { result = step.exit; break; }
    }
    for (const name of ['frontend', 'storybook']) {
      const output = join(scratch, name);
      if (existsSync(output)) receipt.outputs.push({ name, files: inventory(output) });
    }
    if (existsSync(documentation)) {
      const bytes = readFileSync(documentation);
      receipt.generatedDocumentation = { bytes: bytes.length,
        sha256: createHash('sha256').update(bytes).digest('hex') };
    }
    return result;
  } catch (error) {
    receipt.error = { name: error.name, message: error.message };
    throw error;
  } finally {
    // Do not discard a pre-existing file if a step fails or the user interrupts.
    rmSync(documentation, { force: true });
    if (hadDocumentation) renameSync(originalDocumentation, documentation);
    rmSync(scratch, { recursive: true, force: true });
    rmSync(lock, { recursive: true });
    receipt.scratchRemoved = !existsSync(scratch);
    receipt.freeAfterCleanupBytes = freeBytes(root);
    receipt.freeDeltaBytes = receipt.freeAfterCleanupBytes - receipt.freeBeforeBytes;
    receipt.finished = new Date().toISOString();
    save();
  }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const root = resolve(dirname(fileURLToPath(import.meta.url)), '../..');
  const disk = statfsSync(root);
  if (disk.bavail * disk.bsize < 50 * 1024 ** 3) {
    throw new Error('Frontend validation requires at least 50 GiB free host storage.');
  }
  const reports = join(root, 'dist', 'validation');
  mkdirSync(reports, { recursive: true });
  const reportDir = join(reports, `frontend-${new Date().toISOString().replaceAll(':', '-')}-${process.pid}`);
  console.info('Validation evidence:', reportDir);
  process.exitCode = await validate({ root, reportDir });
}
