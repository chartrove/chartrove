import test from 'node:test';
import assert from 'node:assert/strict';
import {MAX_NATIVE_MULTIPART_BYTES, MAX_NATIVE_MULTIPART_PARTS, prepareNativeMultipart} from '../src/assets/desktop/multipart.mjs';

test('native multipart materialization preserves exact file bytes, scalar identity and metadata', async () => {
  const raw = new Uint8Array([0, 255, 13, 10, ...new TextEncoder().encode('{"exact":1.234567890123456789,"text":"café صفر"}')]);
  const form = new FormData();
  form.append('file', new Blob([raw], {type: 'application/fhir+json'}), 'synthetic"-café.json');
  form.append('expected_owner_id', '11111111-1111-4111-8111-111111111111');
  const keys = [...form.keys()];
  const result = await prepareNativeMultipart(form);
  assert.ok(result.body instanceof ArrayBuffer);
  assert.ok(result.contentType.startsWith('multipart/form-data;'));
  const parsed = await new Response(result.body, {headers: {'Content-Type': result.contentType}}).formData();
  assert.equal(parsed.get('expected_owner_id'), form.get('expected_owner_id'));
  const file = parsed.get('file');
  assert.equal(file.type, 'application/fhir+json');
  assert.equal(file.name, form.get('file').name);
  assert.deepEqual(new Uint8Array(await file.arrayBuffer()), raw);
  assert.deepEqual([...form.keys()], keys, 'preparation must not mutate caller intent');
});

test('duplicates remain duplicates for server-side strict rejection', async () => {
  const form = new FormData();
  form.append('expected_owner_id', 'synthetic-first');
  form.append('expected_owner_id', 'synthetic-second');
  const {body, contentType} = await prepareNativeMultipart(form);
  const parsed = await new Response(body, {headers: {'Content-Type': contentType}}).formData();
  assert.deepEqual(parsed.getAll('expected_owner_id'), ['synthetic-first', 'synthetic-second']);
});

test('native multipart rejects excess raw bytes before preparation with a payload-free error', async () => {
  const form = new FormData();
  form.append('file', new Blob([new Uint8Array(MAX_NATIVE_MULTIPART_BYTES + 1)]), 'PRIVATE_FILENAME_CANARY');
  await assert.rejects(prepareNativeMultipart(form), error => {
    assert.equal(error.message, 'Native file request exceeds supported transport limits');
    assert.ok(!error.message.includes('PRIVATE_FILENAME_CANARY'));
    return true;
  });
});

test('native multipart also bounds the encoded wire envelope', async () => {
  const form = new FormData();
  form.append('file', new Blob([new Uint8Array(MAX_NATIVE_MULTIPART_BYTES - 4)]), 'synthetic-envelope.bin');
  await assert.rejects(prepareNativeMultipart(form), /supported transport limits/);
});

test('native multipart bounds field count/name and rejects non-FormData', async () => {
  const excessive = new FormData();
  for (let index = 0; index <= MAX_NATIVE_MULTIPART_PARTS; index++) excessive.append('synthetic-' + index, 'value');
  await assert.rejects(prepareNativeMultipart(excessive), /supported transport limits/);
  const longName = new FormData();
  longName.append('x'.repeat(129), 'synthetic');
  await assert.rejects(prepareNativeMultipart(longName), /supported transport limits/);
  await assert.rejects(prepareNativeMultipart('PRIVATE_BODY_CANARY'), /must be multipart data/);
});
