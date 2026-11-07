import {formatDate} from '@angular/common';

// Keep the precision asserted by a partial FHIR effectiveDateTime. Parsing it
// as a full Date would introduce a January/month-start day that was not supplied.
export function formatObservationDate(date: string | number | Date): string {
  if (typeof date === 'string' && /^\d{4}(-\d{2})?$/.test(date)) {
    return date;
  }
  return date ? formatDate(date, 'mediumDate', 'en-US', undefined) : 'Unknown date';
}
