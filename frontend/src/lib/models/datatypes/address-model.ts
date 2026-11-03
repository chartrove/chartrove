import {evaluateFhirPath} from '../../fhirpath';

export class AddressModel {
  city: string
  line: string[]
  state: string
  country: string
  postalCode: string


  constructor(fhirData: any) {
    if(!fhirData){
      return
    }
    this.city = evaluateFhirPath(fhirData, "city").join("")
    this.line = evaluateFhirPath(fhirData, "line")
    this.state = evaluateFhirPath(fhirData, "state").join("")
    this.country = evaluateFhirPath(fhirData, "country").join("")
    this.postalCode = evaluateFhirPath(fhirData, "postalCode").join("")
  }
}
