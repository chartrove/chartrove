import {LabEditSession, LabOriginal} from './lab-edit-session';

describe('Lab value-only editing session',()=>{
  const source='22222222-2222-4222-8222-222222222222';
  function record(value: string): LabOriginal {
    return {observation:{source_id:source,source_resource_type:'Observation',source_resource_id:'original'} as any,
      raw:`{"resourceType":"Observation","id":"original","code":{"text":"Original label","coding":[{"system":"http://loinc.org","code":"synthetic-code"}]},${value},"extension":[{"url":"urn:synthetic:exact","valueDecimal":1.000000000000000001}],"note":[{"text":"Do not change"}]}`};
  }
  function question(dataType='REAL'):any{return {codeSystem:'http://loinc.org',items:[{question:'Different display label',questionCode:'synthetic-code',questionCodeSystem:'LOINC',dataType}]};}
  function bundle(value:any):any{return {entry:[{resource:{resourceType:'Observation',id:'untrusted-generated',code:{coding:[{system:'http://loinc.org',code:'synthetic-code'}]},...value}}]};}
  function ready(original: LabOriginal, before: any, type='REAL') {
    const session=new LabEditSession(source,[original]);session.prepare(question(type));session.capture(bundle(before));return session;
  }
  for(const [field,type,oldValue,newValue] of [['valueInteger','INT',0,3],['valueBoolean','BL',false,true],['valueString','ST','before','after']] as const) {
    it(`preserves ${field} zero/false values and unrelated precision`,()=>{
      const session=ready(record(`"${field}":${JSON.stringify(oldValue)}`),{[field]:oldValue},type);
      const edits=session.plan(bundle({[field]:newValue}));expect(edits.length).toBe(1);
      expect(edits[0].resourceRawText).toContain('1.000000000000000001');
      expect(JSON.parse(edits[0].resourceRawText)[field]).toBe(newValue);
      expect(JSON.parse(edits[0].resourceRawText).note[0].text).toBe('Do not change');
    });
  }
  it('does not write unchanged answers or recapture an edited value as baseline',()=>{
    const session=ready(record('"valueQuantity":{"value":2.000,"unit":"mg/dL"}'),{valueQuantity:{value:2,unit:'mg/dL'}});
    expect(session.plan(bundle({valueQuantity:{value:2,unit:'mg/dL'}}))).toEqual([]);
    session.capture(bundle({valueQuantity:{value:3,unit:'mg/dL'}}));
    expect(JSON.parse(session.plan(bundle({valueQuantity:{value:3,unit:'mg/dL'}}))[0].resourceRawText).valueQuantity.value).toBe(3);
  });
  it('retains the loaded native tuple if a display model changes while the modal is open',()=>{
    const original=record('"valueString":"before"');
    const session=ready(original,{valueString:'before'},'ST');
    original.observation.source_id='another-source';original.observation.source_resource_id='another-id';
    const edits=session.plan(bundle({valueString:'after'}));
    expect(edits[0].observation.source_id).toBe(source);expect(edits[0].observation.source_resource_id).toBe('original');
  });
  it('refuses unit changes, clearing, type changes, duplicate and unmapped answers',()=>{
    const session=ready(record('"valueQuantity":{"value":2,"unit":"mg/dL"}'),{valueQuantity:{value:2,unit:'mg/dL'}});
    for(const after of [bundle({valueQuantity:{value:3,unit:'g/L'}}),{entry:[]},bundle({valueInteger:3}),{entry:[...bundle({valueQuantity:{value:3,unit:'mg/dL'}}).entry,...bundle({valueQuantity:{value:4,unit:'mg/dL'}}).entry]},
      {entry:[{resource:{resourceType:'Observation',code:{coding:[{system:'http://loinc.org',code:'unknown'}]},valueQuantity:{value:3}}}]}]) {
      expect(()=>session.plan(after)).toThrowError(/safely/);
    }
  });
  it('rejects an unrepresentable existing editable decimal rather than rounding it',()=>{
    expect(()=>new LabEditSession(source,[record('"valueQuantity":{"value":0.1234567890123456789}')])).toThrowError(/safely/);
  });
  it('rejects ambiguous JSON and uses no label-only or foreign-source fallback',()=>{
    const duplicate=record('"valueString":"old","valueString":"another"');
    expect(()=>new LabEditSession(source,[duplicate])).toThrowError(/safely/);
    const original=record('"valueString":"old"');
    expect(()=>new LabEditSession('foreign-source',[original])).toThrowError(/safely/);
    const session=new LabEditSession(source,[original]);
    expect(()=>session.prepare({items:[{question:'Original label',dataType:'ST'}]} as any)).toThrowError(/safely/);
  });
  it('locks unmatched questions and preserves a source unit instead of a panel default',()=>{
    const session=new LabEditSession(source,[record('"valueQuantity":{"value":2,"unit":"original-unit","code":"orig","system":"urn:synthetic:units"}')]);
    const q=question();q.items[0].units=[{name:'panel-default'}];q.items.push({questionCode:'unmatched',questionCodeSystem:'http://loinc.org',dataType:'REAL',defaultAnswer:9,value:9});
    const prepared=session.prepare(q);
    expect(prepared.items[0].unit).toEqual({name:'original-unit',code:'orig',system:'urn:synthetic:units'});
    expect(prepared.items[1].editable).toBe('0');expect(prepared.items[1].defaultAnswer).toBeUndefined();expect(prepared.items[1].value).toBeUndefined();
    expect(q.items[0].units).toEqual([{name:'panel-default'}]);
  });
});
