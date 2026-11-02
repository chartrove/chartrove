import assert from 'node:assert/strict';
import test from 'node:test';
import { restorePrimitiveControls } from '../.storybook/compodoc-compat.ts';

test('Compodoc unknown fields retain controls for literal primitive defaults', () => {
  const defaults = ['false', 'true', '"Chartrove"', "'synthetic'", '-1.25', '1e3'];
  const documentation = {
    components: [{ propertiesClass: defaults.map(defaultValue => ({ type: 'unknown', defaultValue })) }],
  };
  const result = restorePrimitiveControls(documentation);
  assert.deepEqual(result.components[0].propertiesClass.map(member => member.type),
    ['boolean', 'boolean', 'string', 'string', 'number', 'number']);
});

test('known types and nonliteral expressions remain unchanged and are never evaluated', () => {
  const members = [
    { type: 'SomeDomainType', defaultValue: 'false' },
    { type: 'unknown', defaultValue: 'false || throwUnexpected()' },
    { type: 'unknown', defaultValue: '{ enabled: true }' },
    { type: 'unknown', defaultValue: '[]' },
    { type: 'unknown', defaultValue: '`template ${throwUnexpected()}`' },
    { type: 'unknown', defaultValue: 'null' },
    { type: 'unknown' },
  ];
  const documentation = { components: [{ propertiesClass: members }] };
  assert.deepEqual(restorePrimitiveControls(documentation), documentation);
});

test('documentation normalization preserves originals and public member metadata', () => {
  const documentation = { directives: [{
    inputsClass: [{ name: 'enabled', type: 'unknown', defaultValue: 'false', required: true }],
    outputsClass: [{ name: 'changed', type: 'EventEmitter<boolean>' }],
  }], miscellaneous: { marker: 'retain' } };
  const before = structuredClone(documentation);
  const result = restorePrimitiveControls(documentation);
  assert.deepEqual(documentation, before);
  assert.equal(result.directives[0].inputsClass[0].type, 'boolean');
  assert.deepEqual(result.directives[0].outputsClass, before.directives[0].outputsClass);
  assert.deepEqual(result.miscellaneous, before.miscellaneous);
});
