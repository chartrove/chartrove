function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function primitiveType(defaultValue: unknown): string | undefined {
  if (typeof defaultValue !== 'string') return undefined;
  const value = defaultValue.trim();
  if (value === 'true' || value === 'false') return 'boolean';
  if (/^(?:"(?:[^"\\\r\n]|\\.)*"|'(?:[^'\\\r\n]|\\.)*')$/.test(value)) return 'string';
  if (/^-?(?:\d+(?:\.\d*)?|\.\d+)(?:e[+-]?\d+)?$/i.test(value) && Number.isFinite(Number(value))) {
    return 'number';
  }
  return undefined;
}

// Compodoc 2 emits "unknown" for unannotated initialized fields. Storybook then
// skips its default-value inference and loses primitive controls. Recover only
// literal primitive defaults in this documentation copy; never evaluate source.
export function restorePrimitiveControls<T>(documentation: T): T {
  const result = structuredClone(documentation);
  if (!isRecord(result)) return result;
  for (const group of ['components', 'directives', 'pipes', 'injectables', 'classes', 'interfaces']) {
    const entities = result[group];
    if (!Array.isArray(entities)) continue;
    for (const entity of entities) {
      if (!isRecord(entity)) continue;
      for (const key of ['propertiesClass', 'properties', 'inputsClass']) {
        const members = entity[key];
        if (!Array.isArray(members)) continue;
        for (const member of members) {
          if (!isRecord(member) || member['type'] !== 'unknown') continue;
          const type = primitiveType(member['defaultValue']);
          if (type) member['type'] = type;
        }
      }
    }
  }
  return result;
}
