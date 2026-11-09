import {findNodeAtLocation, Node, parseTree, ParseError} from 'jsonc-parser';

export interface ExactObservationPreview {
  readonly value: string;
  readonly comparator: string | null;
  readonly unit: string | null;
  readonly unitCode: string | null;
  readonly effective: string | null;
}

// Read-only presentation, never an import serializer or a clinical calculation.
// Unsupported/ambiguous records retain the caller's exact-JSON fallback.
export function exactObservationPreview(raw: string): ExactObservationPreview | null {
  if (!raw || raw.length > 2 * 1024 * 1024) return null;
  try {
    const errors: ParseError[] = [];
    const root = parseTree(raw, errors, {disallowComments: true, allowTrailingComma: false});
    if (!root || root.type !== 'object' || errors.length) return null;
    // Refuse duplicate fields rather than choosing one parse-tree occurrence.
    const pending: {node: Node; depth: number}[] = [{node: root, depth: 0}];
    let visited = 0;
    while (pending.length) {
      const current = pending.pop();
      if (!current || ++visited > 50000 || current.depth > 128) return null;
      if (current.node.type === 'object') {
        const names = new Set<string>();
        for (const property of current.node.children || []) {
          const name: unknown = property.children?.[0]?.value;
          if (typeof name !== 'string' || names.has(name)) return null;
          names.add(name);
        }
      }
      for (const child of current.node.children || []) pending.push({node: child, depth: current.depth + 1});
    }
    const at = (path: string[]) => findNodeAtLocation(root, path);
    if (at(['resourceType'])?.value !== 'Observation' || at(['component']) || at(['dataAbsentReason'])) return null;
    const choices = (root.children || []).map(property => property.children?.[0]?.value as unknown)
      .filter((name): name is string => typeof name === 'string' && /^value[A-Z]/.test(name));
    if (choices.length !== 1) return null;
    const field = choices[0];
    const node = field === 'valueQuantity' ? at([field, 'value']) : at([field]);
    if (!node) return null;
    let value: string;
    if (field === 'valueQuantity' || field === 'valueInteger') {
      if (node.type !== 'number' || node.length > 256) return null;
      value = raw.slice(node.offset, node.offset + node.length);
      if (field === 'valueInteger' && (!/^-?\d+$/.test(value) || BigInt(value) < -2147483648n || BigInt(value) > 2147483647n)) return null;
    } else if (field === 'valueString' && node.type === 'string' && typeof node.value === 'string' && node.value.length <= 4096) value = node.value;
    else if (field === 'valueBoolean' && node.type === 'boolean') value = node.value === true ? 'true' : 'false';
    else return null;
    const text = (path: string[]): string | null => {
      const candidate = at(path);
      if (!candidate) return null;
      if (candidate.type !== 'string' || typeof candidate.value !== 'string' || candidate.value.length > 256) throw new Error();
      return candidate.value;
    };
    const effectiveChoices = (root.children || []).map(property => property.children?.[0]?.value as unknown)
      .filter((name): name is string => typeof name === 'string' && /^effective[A-Z]/.test(name));
    const effective = effectiveChoices.length === 1 && effectiveChoices[0] === 'effectiveDateTime' ? text(['effectiveDateTime']) : null;
    return {value, comparator: field === 'valueQuantity' ? text([field, 'comparator']) : null,
      unit: field === 'valueQuantity' ? text([field, 'unit']) : null,
      unitCode: field === 'valueQuantity' ? text([field, 'code']) : null, effective};
  } catch { return null; }
}
