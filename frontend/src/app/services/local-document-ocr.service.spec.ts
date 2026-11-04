import {imageDimensions, localOCRAsset, LocalDocumentOCRService, OCR_MAX_FILE_BYTES} from './local-document-ocr.service';

// A tiny synthetic PDF generator keeps acceptance data independent of patients,
// fixtures downloaded at runtime, or another PDF-authoring dependency.
function syntheticPDF(image?: Uint8Array, pages=1): File {
  const encode=(value:string)=>new TextEncoder().encode(value);
  const objects:Uint8Array[]=[];
  const add=(value:string)=>objects.push(encode(value));
  add('<< /Type /Catalog /Pages 2 0 R >>');
  add(`<< /Type /Pages /Count ${pages} /Kids [${Array(pages).fill('3 0 R').join(' ')}] >>`);
  add('<< /Type /Page /Parent 2 0 R /MediaBox [0 0 600 150] /Resources << /Font << /F1 5 0 R >> /XObject << /Im0 6 0 R >> >> /Contents 4 0 R >>');
  const content=image?'q 600 0 0 150 0 0 cm /Im0 Do Q':'BT /F1 22 Tf 20 70 Td (SYNTHETIC LOCAL DOCUMENT) Tj ET';
  add(`<< /Length ${encode(content).length} >>\nstream\n${content}\nendstream`);
  add('<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>');
  if(image) {
    const head=encode(`<< /Type /XObject /Subtype /Image /Width 1200 /Height 300 /ColorSpace /DeviceRGB /BitsPerComponent 8 /Filter /DCTDecode /Length ${image.length} >>\nstream\n`);
    const tail=encode('\nendstream');const stream=new Uint8Array(head.length+image.length+tail.length);
    stream.set(head);stream.set(image,head.length);stream.set(tail,head.length+image.length);objects.push(stream);
  } else add('<< >>');
  const chunks:Uint8Array[]=[encode('%PDF-1.7\n')],offsets=[0];let offset=chunks[0].length;
  objects.forEach((object,index)=>{offsets.push(offset);const head=encode(`${index+1} 0 obj\n`),tail=encode('\nendobj\n');chunks.push(head,object,tail);offset+=head.length+object.length+tail.length;});
  chunks.push(encode(`xref\n0 ${objects.length+1}\n0000000000 65535 f \n${offsets.slice(1).map(value=>String(value).padStart(10,'0')+' 00000 n \n').join('')}trailer\n<< /Size ${objects.length+1} /Root 1 0 R >>\nstartxref\n${offset}\n%%EOF`));
  const bytes=new Uint8Array(chunks.reduce((sum,part)=>sum+part.length,0));let position=0;
  for(const chunk of chunks) {bytes.set(chunk,position);position+=chunk.length;}
  return new File([bytes],'synthetic.pdf',{type:'application/pdf'});
}

async function syntheticImage(type='image/png'):Promise<File> {
  const canvas=document.createElement('canvas');canvas.width=1200;canvas.height=300;
  const context=canvas.getContext('2d');context.fillStyle='white';context.fillRect(0,0,1200,300);
  context.fillStyle='black';context.font='48px Arial';context.fillText('SYNTHETIC LOCAL DOCUMENT',40,150);
  const blob=await new Promise<Blob>(resolve=>canvas.toBlob(resolve,type,0.95));canvas.width=0;canvas.height=0;
  return new File([blob],type==='image/png'?'synthetic.png':'synthetic.jpg',{type});
}

describe('LocalDocumentOCRService real browser engines',()=>{
  let service:LocalDocumentOCRService;
  beforeEach(()=>service=new LocalDocumentOCRService());
  it('keeps asset URLs on the application origin and honors a reverse-proxy base',()=>{
    expect(localOCRAsset('lang/eng.traineddata.gz','https://chartrove.example/web/')).toBe('https://chartrove.example/web/assets/ocr/lang/eng.traineddata.gz');
    expect(()=>localOCRAsset('core/','file:///private/chartrove/')).toThrow();
  });
  it('bounds file size and image dimensions before decoding',async()=>{
    await expectAsync(service.extract(new File([new Uint8Array(OCR_MAX_FILE_BYTES+1)],'synthetic.png'),new AbortController().signal)).toBeRejected();
    const bytes=new Uint8Array(24);bytes.set([137,80,78,71,13,10,26,10,0,0,0,13,73,72,68,82]);
    new DataView(bytes.buffer).setUint32(16,100000);new DataView(bytes.buffer).setUint32(20,100000);
    expect(()=>imageDimensions(bytes)).toThrow();
    expect(()=>imageDimensions(new TextEncoder().encode('not an image'))).toThrow();
  });
  it('extracts a synthetic PNG with bundled English data and WebAssembly',async()=>{
    const urls:string[]=[];const NativeWorker=window.Worker;
    spyOn(window,'Worker').and.callFake((function(url:string|URL,options?:WorkerOptions){
      urls.push(String(url));return new NativeWorker(url,options);
    }) as any);
    const result=await service.extract(await syntheticImage(),new AbortController().signal);
    expect(result.text).toContain('SYNTHETIC LOCAL DOCUMENT');expect(result.pages).toBe(1);expect(result.mediaType).toBe('image/png');
    expect(urls.length).toBeGreaterThan(0);expect(urls.every(url=>new URL(url,document.baseURI).origin===location.origin)).toBeTrue();
    expect(urls.some(url=>url.includes('/assets/ocr/worker-bootstrap.js?job='))).toBeTrue();
  },75000);
  it('extracts embedded PDF text and OCRs a raster PDF page without uploading',async()=>{
    const embedded=await service.extract(syntheticPDF(),new AbortController().signal);
    expect(embedded.text).toContain('SYNTHETIC LOCAL DOCUMENT');expect(embedded.mediaType).toBe('application/pdf');
    const jpeg=new Uint8Array(await (await syntheticImage('image/jpeg')).arrayBuffer());
    const scanned=await service.extract(syntheticPDF(jpeg),new AbortController().signal);
    expect(scanned.text).toContain('SYNTHETIC LOCAL DOCUMENT');expect(scanned.pages).toBe(1);
  },75000);
  it('rejects over-limit and malformed PDFs, and cancellation releases the active slot',async()=>{
    await expectAsync(service.extract(syntheticPDF(undefined,11),new AbortController().signal)).toBeRejected();
    await expectAsync(service.extract(new File(['%PDF-1.7\ninvalid synthetic pdf'],'synthetic.pdf'),new AbortController().signal)).toBeRejected();
    const controller=new AbortController();const file=await syntheticImage();
    const pending=service.extract(file,controller.signal,()=>controller.abort());
    await expectAsync(service.extract(file,new AbortController().signal)).toBeRejected();
    await expectAsync(pending).toBeRejected();
    const result=await service.extract(syntheticPDF(),new AbortController().signal);expect(result.text).toContain('SYNTHETIC');
  },75000);
});
