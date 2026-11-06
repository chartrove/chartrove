import { Component, OnInit, OnDestroy, ChangeDetectionStrategy } from '@angular/core';
import {FastenApiService} from '../../services/fasten-api.service';
import {ResourceFhir} from '../../models/fasten/resource_fhir';
import {evaluateFhirPath} from '../../../lib/fhirpath';
import {Observable, Subscription} from 'rxjs';
import {map} from 'rxjs/operators';
import {ResponseWrapper} from '../../models/response-wrapper';
import {ActivatedRoute, Params} from '@angular/router';
import {FastenDisplayModel} from '../../../lib/models/fasten/fasten-display-model';
import {fhirModelFactory} from '../../../lib/models/factory';
import {ResourceType} from '../../../lib/models/constants';

class ObservationGroup {[key: string]: ResourceFhir[]}
class ObservationGroupInfo {
  observationGroups: ObservationGroup = {}
  observationGroupTitles: {[key: string]: string} = {}
}
class LabResultCodeByDate {
  label: string //lab result coding (system|code)
  value: string //lab result date
}


@Component({
    selector: 'app-report-labs',
    templateUrl: './report-labs.component.html',
    styleUrls: ['./report-labs.component.scss'],
    changeDetection: ChangeDetectionStrategy.Eager,
    standalone: false
})
export class ReportLabsComponent implements OnInit, OnDestroy {
  loading: boolean = false

  currentPage: number = 1 //1-based index due to the way the pagination component works
  pageSize: number = 10
  allObservationGroups: string[] = []


  //diagnostic report data
  reportSourceId: string = ''
  reportResourceType: string = ''
  reportResourceId: string = ''
  reportDisplayModel: FastenDisplayModel = null

  //currentPage data
  observationGroups: ObservationGroup = {}
  observationGroupTitles: {[key: string]: string} = {}

  loadError = false
  private routeRequest?: Subscription
  private groupRequest?: Subscription
  private observationRequest?: Subscription
  private reportsRequest?: Subscription

  isEmptyReport = false

  diagnosticReports: ResourceFhir[] = []

  constructor(
    private fastenApi: FastenApiService,
    private activatedRoute: ActivatedRoute,
  ) { }

  ngOnInit(): void {
    this.loading = true

    this.populateReports()


    //determine if we're requesting all results or just a single report
    //source_id/:resource_type/:resource_id

    this.routeRequest = this.activatedRoute.params.subscribe((routeParams: Params) => {
      this.groupRequest?.unsubscribe()
      this.observationRequest?.unsubscribe()
      this.loading = true
      this.loadError = false
      this.isEmptyReport = false
      this.reportDisplayModel = null
      this.allObservationGroups = []
      this.observationGroups = {}
      this.observationGroupTitles = {}
      this.currentPage = 1
      this.reportSourceId = routeParams['source_id']
      this.reportResourceType = routeParams['resource_type']
      this.reportResourceId = routeParams['resource_id']

      if(this.reportSourceId && this.reportResourceType && this.reportResourceId){
        //we're requesting a single report
        this.groupRequest = this.findLabResultCodesFilteredToReport(this.reportSourceId, this.reportResourceType, this.reportResourceId).subscribe((data) => {
          this.allObservationGroups = data
          this.currentPage = 1 //reset to first page when changing report
          return this.populateObservationsForCurrentPage()
        }, () => this.failReport())
      } else {
        this.groupRequest = this.findLabResultCodesSortedByLatest().subscribe((data) => {
          // this.loading = false
          this.allObservationGroups = data.map((item) => item.label)
          return this.populateObservationsForCurrentPage()
        }, () => this.failReport())
      }
    });

  }

  ngOnDestroy(): void {
    this.routeRequest?.unsubscribe()
    this.groupRequest?.unsubscribe()
    this.observationRequest?.unsubscribe()
    this.reportsRequest?.unsubscribe()
  }

  private failReport(): void {
    this.loading = false
    this.loadError = true
    this.isEmptyReport = false
    this.reportDisplayModel = null
    this.allObservationGroups = []
    this.observationGroups = {}
    this.observationGroupTitles = {}
  }

  //using the current list of allObservationGroups, retrieve a list of observations, group them by observationGroup, and set the observationGroupTitles
  populateObservationsForCurrentPage(){
    this.observationRequest?.unsubscribe()
    this.loadError = false

    let observationGroups = this.allObservationGroups.slice((this.currentPage-1) * this.pageSize, this.currentPage * this.pageSize)

    if (observationGroups.length === 0) {
      this.observationGroups = {}
      this.observationGroupTitles = {}
      this.isEmptyReport = true
      this.loading = false
      return
    }

    this.loading = true
    this.observationRequest = this.getObservationsByCodes(observationGroups).subscribe((data) => {
      this.loading = false
      this.observationGroups = data.observationGroups
      this.observationGroupTitles = data.observationGroupTitles

      this.isEmptyReport = !!!Object.keys(this.observationGroups).length
    }, () => this.failReport())

  }

  //get a list of all lab codes associated with a diagnostic report
  findLabResultCodesFilteredToReport(sourceId: string, resourceType: string, resourceId: string): Observable<string[]> {
    // The importer resolves Bundle fullUrl aliases into source-qualified graph
    // relationships while preserving original FHIR. Never infer IDs from a URI.
    return this.fastenApi.getResourceGraph('MedicalHistory', [{
      source_id: sourceId, source_resource_type: resourceType, source_resource_id: resourceId
    }]).pipe(map(graph => {
      const roots = graph?.results?.['DiagnosticReport'] ?? []
      const report = roots[0]
      if (roots.length !== 1 || resourceType !== 'DiagnosticReport' ||
          report.source_id !== sourceId || report.source_resource_type !== resourceType ||
          report.source_resource_id !== resourceId || report.resource_raw?.resourceType !== 'DiagnosticReport') {
        throw new Error('Selected lab report is unavailable')
      }
      // Direct edges retain DiagnosticReport.result membership. Descendants of
      // those Observations can be derived-from inputs and must not add groups.
      const observations = (report.related_resources ?? []).filter(row => row.source_resource_type === 'Observation')
      const identities = new Set<string>()
      const groups = new Set<string>()
      for (const observation of observations) {
        if (observation.source_id !== sourceId || !observation.source_resource_id || observation.resource_raw?.resourceType !== 'Observation') {
          throw new Error('Selected lab results are unavailable')
        }
        identities.add(observation.source_resource_id)
        const code = evaluateFhirPath(observation.resource_raw, "Observation.code.coding.where(system='http://loinc.org').first().code")[0]
        if (typeof code === 'string' && code.length > 0) groups.add('http://loinc.org|' + code)
      }
      const references = evaluateFhirPath(report.resource_raw, 'DiagnosticReport.result.reference')
      const resultCount = evaluateFhirPath(report.resource_raw, 'DiagnosticReport.result').length
      // Fail closed if relationship metadata cannot account for all referenced
      // results. Ambiguous aliases are an explicit limitation, not an empty report.
      if (references.length !== resultCount || references.some(ref => typeof ref !== 'string' || !ref) ||
          new Set(references).size !== identities.size) {
        throw new Error('Selected lab references are incomplete')
      }
      this.reportDisplayModel = fhirModelFactory(resourceType as ResourceType, report)
      return [...groups]
    }))
  }

  //get a list of all unique lab codes ordered by latest date
  findLabResultCodesSortedByLatest(): Observable<LabResultCodeByDate[]> {
    return this.fastenApi.queryResources({
      select: [],
      from: "Observation",
      where: {
        "code": "http://loinc.org|,urn:oid:2.16.840.1.113883.6.1|",
      },
      aggregations: {
        order_by: {
          field: "sort_date",
          fn: "max"
        },
        group_by: {
          field: "code",
        }
      }
    })
    .pipe(
        map((response: ResponseWrapper) => {
          return response.data as LabResultCodeByDate[]
        }),
    )
  }


  //get a list of the last 10 lab results
  populateReports(){
    this.reportsRequest = this.fastenApi.queryResources({
      select: ["*"],
      from: "DiagnosticReport",
      where: {
        "category": "http://terminology.hl7.org/CodeSystem/v2-0074|LAB",
      },
      limit: 10,
    }).subscribe(results => {
      this.diagnosticReports = results.data
    }, () => { this.diagnosticReports = [] })
  }

  isEmpty(obj: any) {
    return Object.keys(obj).length === 0;
  }

  //private methods

  //get a list of observations that have a matching code
  private getObservationsByCodes(codes: string[]): Observable<ObservationGroupInfo>{
    return this.fastenApi.queryResources({
      select: [],
      from: "Observation",
      where: {
        "code": codes.join(","),
      }
    }).pipe(
      map((response: ResponseWrapper) => {

        let observationGroups: ObservationGroup = {}
        let observationGroupTitles: {[key: string]: string} = {}

        //loop though all observations, group by "code.system": "http://loinc.org"
        for(let observation of response.data){
          let observationGroup = evaluateFhirPath(observation.resource_raw, "Observation.code.coding.where(system='http://loinc.org').first().code")[0]
          observationGroups[observationGroup] = observationGroups[observationGroup] ? observationGroups[observationGroup] : []
          observationGroups[observationGroup].push(observation)

          if(!observationGroupTitles[observationGroup]){
            observationGroupTitles[observationGroup] = evaluateFhirPath(observation.resource_raw, "Observation.code.coding.where(system='http://loinc.org').first().display")[0]
          }
        }

        return {
          observationGroups: observationGroups,
          observationGroupTitles: observationGroupTitles
        }
      })
    );
  }

}
