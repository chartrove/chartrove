import {evaluate} from 'fhirpath';

// Application display/query expressions are local and synchronous.
export function evaluateFhirPath(resource: unknown, expression: string): any[] {
  return evaluate(resource, expression, undefined, undefined, {async: false});
}
