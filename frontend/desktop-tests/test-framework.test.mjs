import assert from 'node:assert/strict';
import { realpathSync } from 'node:fs';
import { createRequire } from 'node:module';
import test from 'node:test';

const projectRequire = createRequire(import.meta.url);
const adapterRequire = createRequire(projectRequire.resolve('karma-jasmine'));
const reporterRequire = createRequire(projectRequire.resolve('karma-jasmine-html-reporter'));

test('console and HTML reporters use the same Jasmine engine installation', () => {
  assert.equal(
    realpathSync(adapterRequire.resolve('jasmine-core')),
    realpathSync(reporterRequire.resolve('jasmine-core')),
    'Reporter selection must not change the test engine',
  );
});

test('Jasmine declarations describe the active browser engine major', () => {
  const engine = adapterRequire('jasmine-core/package.json').version;
  const declarations = projectRequire('@types/jasmine/package.json').version;
  assert.equal(
    declarations.split('.')[0], engine.split('.')[0],
    `Declarations ${declarations} must match the active engine ${engine}`,
  );
});
