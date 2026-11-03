describe('Published LForms lab export contract', () => {
  let frame: HTMLIFrameElement;
  let page: Document;
  let realm: any;

  function loadScript(src: string): Promise<void> {
    return new Promise((resolve, reject) => {
      const script = page.createElement('script');
      script.src = src;
      script.onload = () => resolve();
      script.onerror = () => reject(new Error(`Unable to load local widget asset ${src}`));
      page.head.append(script);
    });
  }

  beforeAll(async () => {
    // Custom elements cannot be unregistered. Keep the real widget private;
    // the existing production-wizard lifecycle suite mocks the global API.
    frame = document.createElement('iframe');
    document.body.append(frame);
    page = frame.contentDocument;
    realm = frame.contentWindow;
    page.open();
    page.write('<!doctype html><html><head><meta charset="utf-8"></head><body><h1>Synthetic lab round-trip</h1><div id="first"></div><div id="second"></div></body></html>');
    page.close();
    await loadScript('/base/node_modules/lforms/dist/lforms/webcomponent/assets/lib/zone.min.js');
    await loadScript('/base/node_modules/lforms/dist/lforms/webcomponent/lhc-forms.js');
    await loadScript('/base/node_modules/lforms/dist/lforms/fhir/lformsFHIRAll.min.js');
    await realm.customElements.whenDefined('wc-lhc-form');
  }, 20000);

  afterAll(() => frame?.remove());

  it('preserves zero, units, Unicode and references across validation and actual input changes', async () => {
    // Create input objects/events inside the widget realm, matching normal browser
    // identity. These are the same public export/container arguments used by both
    // production wizards; their event/submit behavior is tested separately.
    const value = await realm.eval(String.raw`(async()=>{
      const unit={name:'mg/dL',code:'mg/dL',system:'http://unitsofmeasure.org'};
      function model(label,value){return {lformsVersion:'34.0.0',code:'synthetic-'+label,codeSystem:'http://loinc.org',name:'Synthetic '+label,type:'LOINC',items:[{linkId:'/2345-7',questionCode:'2345-7',questionCodeSystem:'http://loinc.org',question:'Synthetic glucose '+label,dataType:'REAL',value,units:[unit],unit,questionCardinality:{min:'1',max:'1'},answerCardinality:{min:'1',max:'1'}},{linkId:'/synthetic-note',questionCode:'synthetic-note',questionCodeSystem:'http://loinc.org',question:'Synthetic note '+label,dataType:'ST',value:'SYNTHETIC café صفر',questionCardinality:{min:'1',max:'1'},answerCardinality:{min:'0',max:'1'}}]};}
      const a=model('first',0), b=model('second',7);const before=JSON.stringify([a,b]);
      async function mount(id,data){const container=document.getElementById(id);const widget=document.createElement('wc-lhc-form');const ready=new Promise((resolve,reject)=>{const timer=setTimeout(()=>reject(Error('No form ready '+id)),10000);widget.addEventListener('onFormReady',()=>{clearTimeout(timer);resolve();},{once:true});widget.addEventListener('onError',e=>{clearTimeout(timer);reject(Error('Form error '+id+': '+JSON.stringify(e.detail)));},{once:true});});widget.options={displayScoreWithAnswerText:false};widget.fhirVersion='R4';widget.questionnaire=data;container.append(widget);await ready;return container;}
      const first=await mount('first',a),second=await mount('second',b);
      const validity=[LForms.Util.checkValidity(first),LForms.Util.checkValidity(second)];
      const exports=[first,second].map(c=>LForms.Util.getFormFHIRData('DiagnosticReport','R4',c,{bundleType:'transaction'}));
      const response=LForms.Util.getFormFHIRData('QuestionnaireResponse','R4',first);
      const input=first.querySelector('input');
      async function enter(value){input.value=value;input.dispatchEvent(new Event('input',{bubbles:true}));input.dispatchEvent(new Event('change',{bubbles:true}));input.dispatchEvent(new Event('blur',{bubbles:true}));await new Promise(r=>setTimeout(r,50));}
      await enter('');const requiredErrors=LForms.Util.checkValidity(first);
      await enter('2.5');const changedValidity=LForms.Util.checkValidity(first);
      const changedExports=[first,second].map(c=>LForms.Util.getFormFHIRData('DiagnosticReport','R4',c,{bundleType:'transaction'}));

      return {title:document.title,href:location.href,bodyText:document.body.innerText,inputs:[...document.querySelectorAll('input')].map(i=>({value:i.value,type:i.type})),validity,exports,response,requiredErrors,changedValidity,changedExports,inputUnchanged:before===JSON.stringify([a,b])};
    })()`);
    expect(value.bodyText).toContain('Synthetic first');
    expect(value.bodyText).toContain('Synthetic second');
    expect(value.validity).toEqual([null, null]);
    expect(value.inputUnchanged).toBeTrue();
    expect(value.requiredErrors?.length).toBeGreaterThan(0);
    expect(value.changedValidity).toBeNull();
    function verify(bundles: any[], values: number[]) {
      for (let i = 0; i < 2; i++) {
        const bundle = bundles[i];
        expect(bundle.resourceType).toBe('Bundle');
        expect(bundle.type).toBe('transaction');
        expect(bundle.entry.length).toBe(3);
        const resources = bundle.entry.map(entry => entry.resource);
        const report = resources.find(resource => resource.resourceType === 'DiagnosticReport');
        const observations = resources.filter(resource => resource.resourceType === 'Observation');
        const quantity = observations.find(resource => resource.code.coding[0].code === '2345-7');
        const note = observations.find(resource => resource.code.coding[0].code === 'synthetic-note');
        console.info('Synthetic widget export runtime', report.meta.tag);
        expect(quantity.valueQuantity).toEqual({value: values[i], unit: 'mg/dL', code: 'mg/dL', system: 'http://unitsofmeasure.org'});
        expect(quantity.code.coding[0].system).toBe('http://loinc.org');
        expect(note.valueString).toBe('SYNTHETIC café صفر');
        expect(report.result.map(result => result.reference).sort())
          .toEqual(observations.map(resource => `Observation/${resource.id}`).sort());
        expect(bundle.entry.every(entry => entry.request.method === 'POST' && entry.request.url === entry.resource.resourceType)).toBeTrue();
      }
      const references = bundles.map(bundle => bundle.entry[0].resource.result.map(result => result.reference));
      expect(references[0].every(reference => !references[1].includes(reference))).toBeTrue();
    }
    verify(value.exports, [0, 7]);
    verify(value.changedExports, [2.5, 7]);
    expect(value.response.item.find(item => item.linkId === '/2345-7').answer[0].valueDecimal).toBe(0);
  });
});
