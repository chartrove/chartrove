/* Lifecycle side channel only. The engine's public API/protocol is unmodified.
 * No document text, image, authorization token or credential enters this channel. */
'use strict';
const job = new URL(self.location.href).searchParams.get('job');
if (!job || !/^[a-f0-9-]{36}$/.test(job)) throw new Error('Invalid OCR job');
const control = new BroadcastChannel(`chartrove-ocr-${job}`);
const stop = () => { control.close(); self.close(); };
control.onmessage = event => { if (event.data === 'cancel') stop(); };
setTimeout(() => { control.postMessage('timeout'); stop(); }, 60000);
control.postMessage('ready');
importScripts('./tesseract/worker.min.js');
