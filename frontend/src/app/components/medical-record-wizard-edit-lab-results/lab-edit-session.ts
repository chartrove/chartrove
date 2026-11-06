import {applyEdits, findNodeAtLocation, modify, Node as JsonNode, parseTree, ParseError} from 'jsonc-parser';
import {ObservationModel} from '../../../lib/models/resources/observation-model';
import {LabresultsQuestionnaire} from '../../models/fasten/labresults-questionnaire';

export interface LabOriginal { observation: ObservationModel; raw: string; }
export interface LabEdit { observation: ObservationModel; resourceRawText: string; }
interface Choice { field: string; value: any; }
interface Binding extends LabOriginal { json: any; choice: Choice; }
const failure = () => new Error('This panel cannot be matched or edited safely. No changes were saved. Inspect the source-specific record.');
const fields = ['valueQuantity', 'valueInteger', 'valueBoolean', 'valueString'];
const key = (o: ObservationModel) => JSON.stringify([o.source_id, o.source_resource_type, o.source_resource_id]);
const system = (s: string) => s === 'LOINC' ? 'http://loinc.org' : s;
function codes(coding: any[]): string[] {
  return (Array.isArray(coding) ? coding : []).filter(c => typeof c?.system === 'string' && c.system && typeof c.code === 'string' && c.code)
    .map(c => JSON.stringify([system(c.system), c.code]));
}
const overlaps = (a: string[], b: string[]) => a.some(c => b.includes(c));
function choice(json: any): Choice {
  const present = Object.keys(json).filter(k => /^value[A-Z]/.test(k));
  if (present.length !== 1 || !fields.includes(present[0]) || json.dataAbsentReason != null) throw failure();
  const field = present[0], value = json[field];
  if (field === 'valueString' && typeof value !== 'string' || field === 'valueBoolean' && typeof value !== 'boolean') throw failure();
  if (field === 'valueInteger' && (!Number.isInteger(value) || value < -2147483648 || value > 2147483647)) throw failure();
  if (field === 'valueQuantity' && (!value || typeof value !== 'object' || value.comparator != null || !Number.isFinite(value.value) || Number.isInteger(value.value) && !Number.isSafeInteger(value.value))) throw failure();
  return {field, value};
}
function strictTree(raw: string): JsonNode {
  const errors: ParseError[] = [];
  const tree = parseTree(raw, errors, {disallowComments:true,allowTrailingComma:false});
  if (!tree || tree.type !== 'object' || errors.length) throw failure();
  function unique(node: JsonNode) {
    if (node.type === 'object') {
      const names = node.children.map(c => c.children[0].value);
      if (new Set(names).size !== names.length) throw failure();
    }
    node.children?.forEach(unique);
  }
  unique(tree); return tree;
}
// Compare decimal spellings without rounding unrelated original JSON numbers.
// Only used to refuse an existing editable value that JS cannot round-trip.
function decimal(raw: string): string {
  const m = /^(-?)(\d+)(?:\.(\d+))?(?:e([+-]?\d+))?$/i.exec(raw);
  if (!m || raw.length > 256) throw failure();
  let digits = (m[2] + (m[3] || '')).replace(/^0+/, '');
  if (!digits) return '0';
  let exponent = BigInt(m[4] || '0') - BigInt((m[3] || '').length);
  while (digits.endsWith('0')) { digits = digits.slice(0,-1); exponent++; }
  return `${m[1]}${digits}e${exponent}`;
}

// One editing session: native tuples are identities, coding identifies a unique
// panel question, labels and generated widget IDs never select records.
export class LabEditSession {
  private records: Binding[];
  private bindings: Binding[] = [];
  private baseline: Map<string, Choice> | null = null;
  constructor(source: string, originals: LabOriginal[]) {
    if (!source || !originals.length || originals.length > 200) throw failure();
    const seen = new Set<string>();
    this.records = originals.map(o => {
      if (o.observation.source_id !== source || o.observation.source_resource_type !== 'Observation' || !o.observation.source_resource_id || seen.has(key(o.observation))) throw failure();
      seen.add(key(o.observation));
      const tree = strictTree(o.raw), json = JSON.parse(o.raw);
      if (json.resourceType !== 'Observation') throw failure();
      const c = choice(json);
      if (c.field === 'valueQuantity' || c.field === 'valueInteger') {
        const node = findNodeAtLocation(tree,c.field === 'valueQuantity' ? [c.field,'value'] : [c.field]);
        if (!node || decimal(o.raw.slice(node.offset,node.offset+node.length)) !== decimal(JSON.stringify(c.field === 'valueQuantity' ? c.value.value : c.value))) throw failure();
      }
      return {...o,observation:{...o.observation},json,choice:c};
    });
  }
  prepare(input: LabresultsQuestionnaire): LabresultsQuestionnaire {
    const q = JSON.parse(JSON.stringify(input));
    const bound = new Set<string>();
    const visit = (items: any[]) => items.forEach(item => {
      if (item.items?.length) { visit(item.items); return; }
      const itemCodes = codes([...(item.codeList || []),{system:item.questionCodeSystem || q.codeSystem,code:item.questionCode}]);
      const matches = this.records.filter(r => overlaps(codes(r.json.code?.coding),itemCodes));
      delete item.value; delete item.defaultAnswer;
      if (!matches.length) { item.editable = '0'; return; }
      if (matches.length !== 1 || bound.has(key(matches[0].observation))) throw failure();
      const r = matches[0], field = r.choice.field;
      const expected = field === 'valueQuantity' ? ['REAL','INT'] : field === 'valueInteger' ? ['INT'] : field === 'valueBoolean' ? ['BL'] : ['ST','TX'];
      if (!expected.includes(item.dataType)) throw failure();
      bound.add(key(r.observation)); this.bindings.push(r);
      item.defaultAnswer = field === 'valueQuantity' ? r.choice.value.value : r.choice.value;
      // Value-only edit: preserve the record's unit, never a panel default or an
      // implicit conversion. The exported unit must remain unchanged at save.
      delete item.unit; item.units = [];
      if (field === 'valueQuantity' && r.choice.value.unit) {
        const v = r.choice.value;
        item.unit = {name:v.unit,...(v.code ? {code:v.code} : {}),...(v.system ? {system:v.system} : {})};
        item.units = [item.unit];
      }
    });
    visit(q.items || []);
    if (!this.bindings.length) throw failure();
    return q;
  }
  private exported(bundle: any): Map<string, Choice> {
    if (!Array.isArray(bundle?.entry)) throw failure();
    const result = new Map<string, Choice>();
    for (const entry of bundle.entry) {
      if (entry?.resource?.resourceType === 'DiagnosticReport') continue;
      if (entry?.resource?.resourceType !== 'Observation') throw failure();
      const matches = this.bindings.filter(r => overlaps(codes(r.json.code?.coding),codes(entry.resource.code?.coding)));
      if (matches.length !== 1 || result.has(key(matches[0].observation))) throw failure();
      result.set(key(matches[0].observation),choice(entry.resource));
    }
    return result;
  }
  capture(bundle: any): void {
    if (this.baseline) return;
    const before = this.exported(bundle);
    for (const r of this.bindings) {
      const c = before.get(key(r.observation));
      if (!c || c.field !== r.choice.field || this.scalar(c) !== this.scalar(r.choice)) throw failure();
    }
    this.baseline = before;
  }
  private scalar(c: Choice): string { return JSON.stringify(c.field === 'valueQuantity' ? c.value.value : c.value); }
  plan(bundle: any): LabEdit[] {
    if (!this.baseline) throw failure();
    const after = this.exported(bundle), edits: LabEdit[] = [];
    for (const r of this.bindings) {
      const id = key(r.observation), before = this.baseline.get(id), next = after.get(id);
      if (!next || next.field !== before.field) throw failure(); // no implicit clearing/type change
      if (next.field === 'valueQuantity' && ['unit','code','system'].some(k => next.value[k] !== before.value[k])) throw failure();
      if (this.scalar(next) === this.scalar(before)) continue;
      const path = next.field === 'valueQuantity' ? [next.field,'value'] : [next.field];
      const raw = applyEdits(r.raw,modify(r.raw,path,next.field === 'valueQuantity' ? next.value.value : next.value,{}));
      edits.push({observation:r.observation,resourceRawText:raw});
    }
    return edits;
  }
}
