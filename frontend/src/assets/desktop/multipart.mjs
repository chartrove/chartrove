// WK custom-scheme requests do not reliably carry streamed FormData. Materialize
// the standard encoder's bytes, not a handwritten MIME representation. Callers
// restrict this adapter to the application's native API, never external services.
export const MAX_NATIVE_MULTIPART_BYTES = 32 * 1024 * 1024;
export const MAX_NATIVE_MULTIPART_PARTS = 16;

export async function prepareNativeMultipart(form) {
  const invalid = () => new Error('Native file request exceeds supported transport limits');
  if (!(form instanceof FormData)) throw new Error('Native file request must be multipart data');
  let parts = 0;
  let bytes = 0;
  const encoder = new TextEncoder();
  for (const [name, value] of form.entries()) {
    if (++parts > MAX_NATIVE_MULTIPART_PARTS || name.length > 128) throw invalid();
    bytes += encoder.encode(name).byteLength;
    if (typeof value === 'string') {
      if (value.length > MAX_NATIVE_MULTIPART_BYTES) throw invalid();
      bytes += encoder.encode(value).byteLength;
    } else {
      if (!(value instanceof Blob) || value.name.length > 4096) throw invalid();
      bytes += value.size;
    }
    if (bytes > MAX_NATIVE_MULTIPART_BYTES) throw invalid();
  }
  const response = new Response(form);
  const contentType = response.headers.get('Content-Type');
  if (!contentType?.startsWith('multipart/form-data;') || !contentType.includes('boundary=')) {
    throw new Error('Native file request could not be prepared');
  }
  if (!response.body) throw new Error('Native file request could not be prepared');
  const reader = response.body.getReader();
  let timedOut = false;
  let overLimit = false;
  const timer = setTimeout(() => {
    timedOut = true;
    void reader.cancel().catch(() => {});
  }, 30000);
  const chunks = [];
  let length = 0;
  try {
    while (true) {
      const {done, value} = await reader.read();
      if (timedOut) throw new Error('Native file preparation timed out');
      if (done) break;
      length += value.byteLength;
      if (length > MAX_NATIVE_MULTIPART_BYTES) {
        overLimit = true;
        throw invalid();
      }
      chunks.push(value);
    }
    const output = new Uint8Array(length);
    let offset = 0;
    for (const chunk of chunks) {
      output.set(chunk, offset);
      offset += chunk.byteLength;
    }
    return {body: output.buffer, contentType};
  } catch {
    void reader.cancel().catch(() => {});
    if (timedOut) throw new Error('Native file preparation timed out');
    if (overLimit) throw invalid();
    throw new Error('Native file request could not be prepared');
  } finally {
    clearTimeout(timer);
    reader.releaseLock();
  }
}
