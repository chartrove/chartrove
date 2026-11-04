import {ReferenceModel} from '../datatypes/reference-model';
import {fhirVersions, ResourceType} from '../constants';
import {FastenOptions} from '../fasten/fasten-options';
import {FastenDisplayModel} from '../fasten/fasten-display-model';

export class BinaryModel extends FastenDisplayModel{

  content_type: string | undefined
  content: string | undefined  //decoded data
  data: string | undefined //raw base64 encoded data

  constructor(fhirResource: any, fhirVersion?: fhirVersions, fastenOptions?: FastenOptions) {
    super(fastenOptions)
    this.source_resource_type = ResourceType.Binary

    this.content = fhirResource.data ? atob(fhirResource.data) : '';
    this.data = fhirResource.data
    this.content_type = fhirResource.contentType
    if (this.content_type === 'text/plain' && this.content) {
      // FHIR attachments carry bytes. Reviewed UTF-8 transcripts must not be
      // displayed as atob's Latin-1 byte string; binary image/PDF data stays intact.
      try {
        this.content = new TextDecoder('utf-8', {fatal: true}).decode(Uint8Array.from(this.content, char => char.charCodeAt(0)));
      } catch { /* Keep legacy non-UTF-8 plain text readable as before. */ }
    }
  }
}
