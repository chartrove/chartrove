import { Injectable } from '@angular/core';
import type { Worker as TesseractWorker } from 'tesseract.js';
import type { PDFDocumentLoadingTask, RenderTask } from 'pdfjs-dist';
import { uuidV4 } from '../../lib/utils/uuid';

export const OCR_MAX_FILE_BYTES = 8 * 1024 * 1024;
export const OCR_MAX_PAGES = 10;
export const OCR_MAX_PIXELS = 8 * 1024 * 1024;
export const OCR_MAX_INPUT_PIXELS = 24 * 1024 * 1024;
export const OCR_MAX_TEXT_BYTES = 128 * 1024;
export interface LocalExtraction { text: string; pages: number; mediaType: string }

export function localOCRAsset(path: string, base = document.baseURI): string {
  const url = new URL(`assets/ocr/${path}`, base);
  if (url.origin !== new URL(base).origin || !['http:', 'https:'].includes(url.protocol)) throw new Error('Local OCR requires a web origin');
  return url.href;
}

function checkPixels(width: number, height: number, limit = OCR_MAX_PIXELS): void {
  if (!Number.isSafeInteger(width) || !Number.isSafeInteger(height) || width < 1 || height < 1 || width > 16384 || height > 16384 || width * height > limit) throw new Error('Image dimensions exceed the local processing limit');
}

// Bound dimensions before asking the browser to decode untrusted image pixels.
// Only PNG's mandatory IHDR and JPEG's frame headers are accepted.
export function imageDimensions(bytes: Uint8Array): {width: number; height: number} {
  const view = new DataView(bytes.buffer, bytes.byteOffset, bytes.byteLength);
  if (bytes.length >= 24 && bytes.slice(0,8).every((value,i) => value === [137,80,78,71,13,10,26,10][i]) && view.getUint32(8) === 13 && String.fromCharCode(...bytes.slice(12,16)) === 'IHDR') {
    const width = view.getUint32(16), height = view.getUint32(20); checkPixels(width,height,OCR_MAX_INPUT_PIXELS); return {width,height};
  }
  if (bytes[0] === 255 && bytes[1] === 216) {
    let offset = 2;
    while (offset + 4 < bytes.length) {
      if (bytes[offset++] !== 255) break;
      while (bytes[offset] === 255) offset++;
      const marker = bytes[offset++];
      if (marker === 218 || marker === 217 || offset + 2 > bytes.length) break;
      if (marker === 1 || (marker >= 208 && marker <= 215)) continue;
      const length = view.getUint16(offset);
      if (length < 2 || offset + length > bytes.length) break;
      if ([192,193,194,195,197,198,199,201,202,203,205,206,207].includes(marker) && length >= 8) {
        const height = view.getUint16(offset+3), width = view.getUint16(offset+5); checkPixels(width,height,OCR_MAX_INPUT_PIXELS); return {width,height};
      }
      offset += length;
    }
  }
  throw new Error('Unsupported or invalid PNG/JPEG image');
}

function abortable<T>(promise: Promise<T>, signal: AbortSignal): Promise<T> {
  if (signal.aborted) return Promise.reject(new Error('Local processing cancelled'));
  return new Promise<T>((resolve,reject) => {
    const abort = () => reject(new Error('Local processing cancelled'));
    signal.addEventListener('abort',abort,{once:true});
    promise.then(resolve,reject).finally(() => signal.removeEventListener('abort',abort));
  });
}

@Injectable({providedIn:'root'})
export class LocalDocumentOCRService {
  private active = false;

  async extract(file: File, signal: AbortSignal, progress: (page: number, pages: number) => void = () => {}): Promise<LocalExtraction> {
    if (this.active) throw new Error('Another local document is processing');
    if (!file.size || file.size > OCR_MAX_FILE_BYTES) throw new Error('Choose a document no larger than 8 MiB');
    if (typeof BroadcastChannel === 'undefined' || typeof Worker === 'undefined' || typeof WebAssembly === 'undefined') throw new Error('Local processing requires a current web browser');
    this.active = true;
    const controller = new AbortController();
    const relay = () => controller.abort(); signal.addEventListener('abort',relay,{once:true}); if (signal.aborted) relay();
    const deadline = setTimeout(() => controller.abort(),60000);
    let worker: TesseractWorker | undefined;
    let loading: PDFDocumentLoadingTask | undefined;
    let rendering: RenderTask | undefined;
    let channel: BroadcastChannel | undefined;
    let channelReady = false;
    let channelClosing = false;
    let channelDeadline: ReturnType<typeof setTimeout> | undefined;
    const closeChannel = () => { channel?.postMessage('cancel'); channel?.close(); channel = undefined; if (channelDeadline) clearTimeout(channelDeadline); };
    const stop = () => { rendering?.cancel(); void loading?.destroy().catch(() => {}); if (worker) void worker.terminate().catch(() => {}); channel?.postMessage('cancel'); };
    controller.signal.addEventListener('abort',stop,{once:true});
    const getWorker = async (): Promise<TesseractWorker> => {
      if (worker) return worker;
      const tesseract = await abortable(import('tesseract.js'),controller.signal);
      const job = uuidV4();
      channel = new BroadcastChannel(`chartrove-ocr-${job}`);
      channel.onmessage = event => {
        if (event.data === 'ready') { channelReady = true; if (channelClosing || controller.signal.aborted) closeChannel(); }
        if (event.data === 'timeout') controller.abort();
      };
      const workerURL = new URL(localOCRAsset('worker-bootstrap.js')); workerURL.searchParams.set('job',job);
      const opening = tesseract.createWorker('eng',tesseract.OEM.LSTM_ONLY,{
        workerPath:workerURL.href, corePath:localOCRAsset('core/'), langPath:localOCRAsset('lang/'),
        workerBlobURL:false, cacheMethod:'none', gzip:true, logger:() => {}, errorHandler:() => controller.abort()
      });
      opening.then(created => { if (controller.signal.aborted || channelClosing) void created.terminate(); else worker = created; },() => {});
      worker = await abortable(opening,controller.signal);
      return worker;
    };
    const recognize = async (canvas: HTMLCanvasElement): Promise<string> => {
      const engine = await getWorker();
      const result = await abortable(engine.recognize(canvas,{}, {text:true}),controller.signal);
      return result.data.text;
    };
    const pages: string[] = [];
    let mediaType = '';
    const append = (text: string) => {
      pages.push(text);
      if (new TextEncoder().encode(pages.join('\n\n')).length > OCR_MAX_TEXT_BYTES) throw new Error('Extracted text exceeds 128 KiB');
    };
    try {
      const header = new Uint8Array(await abortable(file.slice(0,256*1024).arrayBuffer(),controller.signal));
      if (String.fromCharCode(...header.slice(0,5)) === '%PDF-') {
        mediaType = 'application/pdf';
        const pdf = await abortable(import('pdfjs-dist'),controller.signal);
        pdf.GlobalWorkerOptions.workerSrc = localOCRAsset('pdfjs/pdf.worker.min.mjs');
        loading = pdf.getDocument({
          data:new Uint8Array(await abortable(file.arrayBuffer(),controller.signal)),
          cMapUrl:localOCRAsset('pdfjs/cmaps/'),cMapPacked:true,standardFontDataUrl:localOCRAsset('pdfjs/fonts/'),
          wasmUrl:localOCRAsset('pdfjs/wasm/'),iccUrl:localOCRAsset('pdfjs/iccs/'),
          useWorkerFetch:true,stopAtErrors:true,maxImageSize:OCR_MAX_INPUT_PIXELS,canvasMaxAreaInBytes:OCR_MAX_PIXELS*4,
          disableAutoFetch:true,disableStream:true,disableRange:true,enableXfa:false,verbosity:0
        });
        const document = await abortable(loading.promise,controller.signal);
        if (document.numPages < 1 || document.numPages > OCR_MAX_PAGES) throw new Error('PDFs must have at most 10 pages');
        for (let pageNumber=1; pageNumber<=document.numPages; pageNumber++) {
          progress(pageNumber,document.numPages);
          const page = await abortable(document.getPage(pageNumber),controller.signal);
          try {
            const content = await abortable(page.getTextContent(),controller.signal);
            const embedded = content.items.map(item => 'str' in item ? item.str + (item.hasEOL ? '\n' : ' ') : '').join('').trim();
            if (embedded) append(embedded);
            else {
              const viewport = page.getViewport({scale:1.5}); checkPixels(Math.ceil(viewport.width),Math.ceil(viewport.height));
              const canvas = globalThis.document.createElement('canvas'); canvas.width = Math.ceil(viewport.width); canvas.height = Math.ceil(viewport.height);
              try { rendering = page.render({canvas,viewport}); await abortable(rendering.promise,controller.signal); rendering = undefined; append(await recognize(canvas)); }
              finally { canvas.width = 0; canvas.height = 0; }
            }
          } finally { page.cleanup(); }
        }
      } else {
        mediaType = header[0] === 137 ? 'image/png' : 'image/jpeg';
        const expected = imageDimensions(header);
        const decoding = createImageBitmap(file);
        decoding.then(bitmap => { if (controller.signal.aborted) bitmap.close(); },() => {});
        const bitmap = await abortable(decoding,controller.signal);
        const canvas = document.createElement('canvas');
        try {
          checkPixels(bitmap.width,bitmap.height,OCR_MAX_INPUT_PIXELS);
          if (!((bitmap.width === expected.width && bitmap.height === expected.height) || (bitmap.width === expected.height && bitmap.height === expected.width))) throw new Error('Image dimensions changed during decoding');
          const scale = Math.min(1,Math.sqrt(OCR_MAX_PIXELS/(bitmap.width*bitmap.height)));
          canvas.width = Math.max(1,Math.floor(bitmap.width*scale)); canvas.height = Math.max(1,Math.floor(bitmap.height*scale));
          const context = canvas.getContext('2d'); if (!context) throw new Error('Canvas is unavailable');
          context.drawImage(bitmap,0,0,canvas.width,canvas.height); progress(1,1); append(await recognize(canvas));
        } finally { bitmap.close(); canvas.width = 0; canvas.height = 0; }
      }
      if (controller.signal.aborted) throw new Error('Local processing cancelled');
      const text = pages.join('\n\n'); if (!text.trim()) throw new Error('No readable text was found');
      return {text,pages:pages.length,mediaType};
    } finally {
      clearTimeout(deadline); signal.removeEventListener('abort',relay); controller.signal.removeEventListener('abort',stop);
      channelClosing = true;
      stop();
      if (channelReady) closeChannel();
      else if (channel) channelDeadline = setTimeout(closeChannel,60000);
      this.active = false;
    }
  }
}
