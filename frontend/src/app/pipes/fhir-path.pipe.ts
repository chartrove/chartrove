import { Pipe, PipeTransform } from '@angular/core';
import {ResourceFhir} from '../models/fasten/resource_fhir';
import {evaluateFhirPath} from '../../lib/fhirpath';

@Pipe({
    name: 'fhirPath',
    standalone: false
})
export class FhirPathPipe implements PipeTransform {

  transform(resourceFhir: ResourceFhir, ...pathQueryWithFallback: string[]): string {
    if(!resourceFhir){
      return null
    }

    for(let pathQuery of pathQueryWithFallback){
      let result = evaluateFhirPath(resourceFhir.resource_raw, pathQuery).join(", ")
      if(result){
        return result
      }
    }

    return null
  }

}
