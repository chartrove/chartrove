import {TestBed} from '@angular/core/testing';
import {RouterTestingModule} from '@angular/router/testing';
import {ObservationComponent} from './observation.component';
import {ObservationModel} from '../../../../../lib/models/resources/observation-model';

describe('Observation clinically relevant date versus issuance (#440)',()=>{
  beforeEach(async()=>{
    await TestBed.configureTestingModule({imports:[ObservationComponent,RouterTestingModule]}).compileComponents();
  });
  for(const scenario of [
    {name:'distinct effective and issued values',effective:'2020-02-01',issued:'2020-02-03T10:30:00+05:30'},
    {name:'effective date without issuance',effective:'2020-02-01',issued:undefined},
    {name:'issuance without an effective date',effective:undefined,issued:'2020-02-03T10:30:00Z'},
    {name:'neither asserted date',effective:undefined,issued:undefined},
    {name:'year precision without an inferred month or day',effective:'2020',issued:undefined},
    {name:'month precision without an inferred day',effective:'2020-02',issued:undefined},
  ]) {
    it(`preserves ${scenario.name} under the correct visible labels`,()=>{
      const raw={resourceType:'Observation',id:'synthetic-observation-dates',status:'final',code:{text:'Synthetic result'},valueString:'Synthetic café صفر',
        ...(scenario.effective ? {effectiveDateTime:scenario.effective}:{}),...(scenario.issued ? {issued:scenario.issued}:{}),note:[{text:'Preserve source assertion'}]};
      const before=JSON.stringify(raw),model=new ObservationModel(raw);
      const fixture=TestBed.createComponent(ObservationComponent);fixture.componentInstance.displayModel=model;fixture.detectChanges();
      const rows=Array.from(fixture.nativeElement.querySelectorAll('fhir-ui-table tr') as NodeListOf<HTMLTableRowElement>);
      const visible=new Map(rows.map(row=>[row.cells[0].textContent.trim(),row.cells[1].textContent.trim()]));
      expect(visible.get('Issued on')).toBe(scenario.issued);
      expect(visible.get('Observed on')).toBe(scenario.effective);
      expect(visible.get('Value')).toBe('Synthetic café صفر');
      expect(model.effective_date).toBe(scenario.effective);
      expect(JSON.stringify(raw)).toBe(before);
    });
  }
});
