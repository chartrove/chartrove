import {Component, EventEmitter, OnDestroy, OnInit, Optional, Output, ViewChild, ChangeDetectionStrategy} from '@angular/core';
import {Patient} from 'fhir/r4';
import {EpicSandboxService} from '../../services/epic-sandbox.service';
import {MedicalSourcesConnectedComponent} from '../../components/medical-sources-connected/medical-sources-connected.component';
import {LighthouseService} from '../../services/lighthouse.service';
import {FastenApiService} from '../../services/fasten-api.service';
import {LighthouseSourceMetadata} from '../../models/lighthouse/lighthouse-source-metadata';
import {Source} from '../../models/fasten/source';
import {NgbModal, type NgbModalRef} from '@ng-bootstrap/ng-bootstrap';
import {ActivatedRoute} from '@angular/router';
import {environment} from '../../../environments/environment';
import {BehaviorSubject, firstValueFrom, forkJoin, Observable, of, Subject} from 'rxjs';
import {
  LighthouseSourceSearch,
  LighthouseSourceSearchAggregation,
  LighthouseBrandListDisplayItem
} from '../../models/lighthouse/lighthouse-source-search';
import {debounceTime, distinctUntilChanged, pairwise, startWith} from 'rxjs/operators';
import {MedicalSourcesFilter, MedicalSourcesFilterService} from '../../services/medical-sources-filter.service';
import {FormControl, FormGroup} from '@angular/forms';
import * as _ from 'lodash';
import {PatientAccessBrand} from '../../models/patient-access-brands';
import {PlatformService} from '../../services/platform.service';
import {FormRequestHealthSystemComponent} from '../../components/form-request-health-system/form-request-health-system.component';
import {extractErrorFromResponse} from '../../../lib/utils/error_extract';
import {AuthService} from '../../services/auth.service';

export const sourceConnectWindowTimeout = 24*5000 //wait 2 minutes (5 * 24 = 120)

interface ManualFileSelection {
  readonly addedFiles?: readonly File[];
  readonly rejectedFiles?: readonly unknown[];
}

export class SourceListItem {
  source?: Source
  brand: LighthouseBrandListDisplayItem | PatientAccessBrand
  searchHighlights?: string[]
}

@Component({
    selector: 'app-medical-sources',
    templateUrl: './medical-sources.component.html',
    styleUrls: ['./medical-sources.component.scss'],
    changeDetection: ChangeDetectionStrategy.Eager,
    standalone: false
})
export class MedicalSourcesComponent implements OnInit, OnDestroy {
  loading: boolean = false

  environment_name = environment.environment_name

  sandboxStatus = '';
  sandboxError = '';
  sandboxPatient?: Patient;
  sandboxImportedSource?: Source;
  sandboxImporting = false;
  sandboxController?: AbortController;
  private destroyed = false;
  @ViewChild(MedicalSourcesConnectedComponent) connectedSources?: MedicalSourcesConnectedComponent;

  uploadingBundle = false
  uploadedFile: File[] = []
  uploadErrorMsg: string = ""

  searchTermUpdate = new BehaviorSubject<string>("");
  status: { [name: string]: undefined | "token" | "authorize" } = {}

  //aggregation/filter data & limits
  globalLimits: {
    // aggregations: LighthouseSourceSearchAggregations | undefined,
  } = {
    // categories: [],
    // aggregations: undefined,
  }




  //source of truth for current state
  //TODO: see if we can remove this without breaking search/filtering
  filterForm = this.filterService.filterForm;

  //modal
  modalSelectedBrandListItem: LighthouseBrandListDisplayItem | PatientAccessBrand = null;
  modalCloseResult = '';


  // CCDA-FHIR modal
  @ViewChild('ccdaWarningModalRef') ccdaWarningModalRef : any;
  private ccdaWarningModal?: NgbModalRef;

  constructor(
    private lighthouseApi: LighthouseService,
    private fastenApi: FastenApiService,
    private platformApi: PlatformService,
    private activatedRoute: ActivatedRoute,
    private filterService: MedicalSourcesFilterService,
    private modalService: NgbModal,
    private epicSandbox: EpicSandboxService,
    private auth: AuthService,

  ) {
  }

  ngOnInit(): void {

  }

  async connectEpicSandbox() {
    if (this.sandboxController || this.sandboxImporting) return;
    this.sandboxPatient = undefined;
    this.sandboxImportedSource = undefined;
    this.sandboxError = '';
    const controller = this.sandboxController = new AbortController();
    try {
      const patient = await this.epicSandbox.connect(controller.signal, message => {
        if (!this.destroyed) this.sandboxStatus = message;
      });
      if (!this.destroyed && !controller.signal.aborted) {
        this.sandboxPatient = patient;
        this.sandboxStatus = 'Connected to Epic sandbox. The synthetic Patient is ready to import.';
      }
    } catch (error) {
      if (!this.destroyed) {
        if (controller.signal.aborted) this.sandboxStatus = 'Connection cancelled.';
        else { this.sandboxStatus = ''; this.sandboxError = error.message; }
      }
    } finally {
      if (this.sandboxController === controller) this.sandboxController = undefined;
    }
  }

  cancelEpicSandbox() { this.sandboxController?.abort(); }

  async importEpicSandboxPatient() {
    if (!this.sandboxPatient || this.sandboxImporting) return;
    this.sandboxImporting = true;
    this.sandboxError = '';
    try {
      const result = await this.fastenApi.importSandboxPatient(this.sandboxPatient).toPromise();
      if (!this.destroyed) {
        this.sandboxImportedSource = result.source;
        this.sandboxPatient = undefined;
        this.sandboxStatus = result.alreadyImported
          ? 'This synthetic Patient is already imported. No existing records were changed.'
          : 'Synthetic Patient imported. Reconnect to retrieve it again; automatic sync is not enabled.';
        this.connectedSources?.refreshSources();
      }
    } catch (error) {
      if (!this.destroyed) this.sandboxError = 'The local import failed. You can retry without signing in again.';
    } finally {
      this.sandboxImporting = false;
    }
  }

  ngOnDestroy() {
    this.destroyed = true;
    this.sandboxController?.abort();
    this.sandboxPatient = undefined;
    this.uploadedFile = [];
    this.uploadErrorMsg = '';
    this.ccdaWarningModal?.dismiss('receiving page closed');
    this.ccdaWarningModal = undefined;
  }



  //OLD FUNCTIONS
  //
  //
  // private populateAvailableSourceList(results: LighthouseSourceSearch): void {
  //   console.log("AGGREGATIONS!!!!!", results.aggregations)
  //   this.totalAvailableSourceList = results.hits.total.value
  //   if(results.hits.hits.length == 0){
  //     this.scrollComplete = true
  //     console.log("scroll complete")
  //     return
  //   }
  //   this.scrollId = results._scroll_id
  //   this.availableSourceList = this.availableSourceList.concat(results.hits.hits.map((result) => {
  //     return {metadata: result._source}
  //   }).filter((item) => {
  //     return !this.connectedSourceList.find((connectedItem) => connectedItem.metadata.source_type == item.metadata.source_type)
  //   }))
  // }
  //


  // /**
  //  * after pressing the logo (connectModalHandler button), this function will display a modal with information about the source
  //  * @param $event
  //  * @param sourceType
  //  */
  public connectModalHandler(contentModalRef, sourceListItem: SourceListItem) :void {
    console.log("TODO: connect Handler")


    this.modalSelectedBrandListItem = sourceListItem.brand
    this.modalService.open(contentModalRef, {ariaLabelledBy: 'modal-basic-title'}).result.then((result) => {
      this.modalSelectedBrandListItem = null
      this.modalCloseResult = `Closed with: ${result}`;
    }, (reason) => {
      this.modalSelectedBrandListItem = null
    });
  }

  // /**
  //  * after pressing the connect button in the Modal, this function will generate an authorize url for this source, and redirect the user.
  //  * @param $event
  //  * @param sourceType
  //  */
  public connectHandler($event, brandId: string, portalId: string, endpointId: string): void {

    ($event.currentTarget as HTMLButtonElement).disabled = true;
    this.status[brandId] = "authorize"
    this.status[endpointId] = "authorize"

    this.lighthouseApi.getLighthouseSource(endpointId)
      .then(async (sourceMetadata: LighthouseSourceMetadata) => {
        sourceMetadata.brand_id = brandId
        sourceMetadata.portal_id = portalId

        let authorizationUrl = await this.lighthouseApi.generateSourceAuthorizeUrl(sourceMetadata)

        // redirect to lighthouse with uri's (or open a new window in desktop mode)
        this.lighthouseApi.redirectWithOriginAndDestination(authorizationUrl.toString(), sourceMetadata).subscribe((desktopRedirectData) => {
          if(!desktopRedirectData){
            return //wait for redirect
          }

          //Note: this code will only run in Desktop mode (with popups)
          //in non-desktop environments, the user is redirected in the same window, and this code is never executed.

          //always close the modal
          this.modalService.dismissAll()

          //redirect the browser back to this page with the code in the query string parameters
          this.lighthouseApi.redirectWithDesktopCode(desktopRedirectData.state, desktopRedirectData.codeData)
        })
      });
  }



  private receivingOwnerID(): string | undefined {
    try {
      const id = this.auth.GetCurrentUser()?.native_user_id;
      if (typeof id === 'string' && /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(id) &&
          id !== '00000000-0000-0000-0000-000000000000') return id.toLowerCase();
    } catch { /* Missing/malformed browser identity cannot select a receiving profile. */ }
    return undefined;
  }

  /** Preserve the selected immutable receiver across confirmation, conversion and import. */
  public async uploadSourceBundleHandler(event: ManualFileSelection | null | undefined): Promise<void> {
    if (this.destroyed || this.uploadingBundle || !event?.addedFiles?.length) return;

    this.uploadErrorMsg = '';
    const receivingOwnerID = this.receivingOwnerID();
    if (!receivingOwnerID) {
      this.uploadErrorMsg = 'Sign in again before choosing a file for import.';
      this.uploadedFile = [];
      return;
    }
    let processingFile = event.addedFiles[0];
    this.uploadedFile = [processingFile];
    this.uploadingBundle = true;
    let stage = 'uploading';
    let conversionSent = false;
    let importSent = false;
    const sameReceiver = (): boolean => {
      if (this.destroyed) return false;
      if (this.receivingOwnerID() === receivingOwnerID) return true;
      this.uploadedFile = [];
      this.uploadErrorMsg = importSent
        ? 'Receiving profile changed. The import may have been accepted for the original profile. Check it before retrying.'
        : conversionSent
          ? 'Receiving profile changed. A conversion request may have been sent; no local import was started. Choose the intended profile and select the file again.'
          : 'Receiving profile changed before submission. Choose the intended profile and select the file again.';
      return false;
    };

    try {
      if (processingFile.type === 'text/xml') {
        stage = 'converting';
        const shouldConvert = await this.showCcdaWarningModal();
        if (!shouldConvert || !sameReceiver()) return;
        conversionSent = true;
        processingFile = await firstValueFrom(this.platformApi.convertCcdaToFhir(processingFile));
      }

      if (!sameReceiver()) return;
      stage = 'uploading';
      importSent = true;
      await firstValueFrom(this.fastenApi.createManualSource(processingFile, receivingOwnerID));
      sameReceiver(); // An earlier response must not populate errors/data under a replacement profile.
    } catch (err: unknown) {
      if (sameReceiver()) this.uploadErrorMsg = `Error ${stage} file: ` + (extractErrorFromResponse(err) || 'Unknown Error');
    } finally {
      this.uploadingBundle = false;
      this.uploadedFile = [];
    }
  }

  showCcdaWarningModal(): Promise<boolean> {
    const modal = this.modalService.open(this.ccdaWarningModalRef);
    this.ccdaWarningModal = modal;
    return modal.result.then(() => true, () => false).finally(() => {
      if (this.ccdaWarningModal === modal) this.ccdaWarningModal = undefined;
    });
  }

}
