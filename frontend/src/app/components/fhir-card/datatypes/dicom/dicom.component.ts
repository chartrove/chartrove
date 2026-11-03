import {AfterViewInit, ChangeDetectionStrategy, Component, Input, OnDestroy, TemplateRef} from '@angular/core';
import * as dwv from 'dwv';
import {NgbModal, NgbModalModule, NgbPaginationModule, NgbTooltipModule} from '@ng-bootstrap/ng-bootstrap';
import {BinaryModel} from '../../../../../lib/models/resources/binary-model';

// Adapted from https://raw.githubusercontent.com/ivmartel/dwv-angular/master/src/app/dwv/dwv.component.ts
let nextViewerId = 0;

@Component({
  imports: [NgbModalModule, NgbTooltipModule, NgbPaginationModule],
  providers: [NgbModalModule],
  selector: 'fhir-dicom',
  templateUrl: './dicom.component.html',
  changeDetection: ChangeDetectionStrategy.Eager,
  styleUrls: ['./dicom.component.scss']
})
export class DicomComponent implements AfterViewInit, OnDestroy {
  @Input() displayModel: BinaryModel;

  readonly layerGroupId = `layerGroup${nextViewerId++}`;
  loadProgress = 0;
  dataLoaded = false;
  loadError = '';
  readonly tools = {
    Scroll: new dwv.ToolConfig(),
    ZoomAndPan: new dwv.ToolConfig(),
    WindowLevel: new dwv.ToolConfig(),
    Draw: new dwv.ToolConfig(['Ruler'])
  };
  selectedTool = 'Scroll';

  private dwvApp: dwv.App;
  private initialized = false;
  private destroyed = false;
  private resizeHandler: () => void;
  private listeners: Array<{type: string; callback: (event: any) => void}> = [];
  private orientation: 'axial' | 'coronal' | 'sagittal';
  metaData: Array<{name: string; value: unknown}> = [];
  visibleMetaData: Array<{id: number; name: string; value: unknown}> = [];
  tagsPage = 1;
  tagsPageSize = 10;

  constructor(private modalService: NgbModal) {}

  ngAfterViewInit(): void {
    // The layer container must exist before the library creates its views.
    try {
      this.dwvApp = new dwv.App();
      const options = new dwv.AppOptions({'*': [new dwv.ViewConfig(this.layerGroupId)]});
      options.tools = this.tools;
      this.dwvApp.init(options);
      this.initialized = true;

      let firstRender = true;
      this.listen('loadstart', () => {
        this.dataLoaded = false;
        this.loadProgress = 0;
        this.loadError = '';
        this.metaData = [];
        this.visibleMetaData = [];
        this.tagsPage = 1;
        firstRender = true;
      });
      this.listen('loadprogress', event => {this.loadProgress = event.loaded;});
      this.listen('renderend', () => {
        if (firstRender) {
          firstRender = false;
          const tool = this.dwvApp.canScroll() ? 'Scroll' : 'ZoomAndPan';
          // Rendering can precede the final load event.
          this.selectedTool = tool;
          this.dwvApp.setTool(tool);
        }
      });
      this.listen('load', event => {
        const id = event.dataid ?? this.dwvApp.getDataIds()[0];
        const elements = dwv.getAsSimpleElements(this.dwvApp.getMetaData(id));
        this.metaData = Object.entries(elements).map(([name, value]) => ({name, value}));
        this.dataLoaded = true;
      });
      this.listen('error', () => this.failLoad());
      this.listen('abort', () => this.failLoad());
      this.listen('keydown', event => this.dwvApp.defaultOnKeydown(event));
      this.resizeHandler = this.dwvApp.onResize;
      window.addEventListener('resize', this.resizeHandler);

      if (this.displayModel) {
        const blob = this.dataBase64toBlob(this.displayModel.data, 'application/dicom');
        this.dwvApp.loadFiles([new File([blob], 'dicom.dcm', {type: 'application/dicom'})]);
      }
    } catch {
      // Publish synchronous initialization errors after Angular's first view check.
      queueMicrotask(() => {if (!this.destroyed) this.failLoad();});
    }
  }

  private listen(type: string, callback: (event: any) => void): void {
    const guarded = (event: any) => {if (!this.destroyed) callback(event);};
    this.listeners.push({type, callback: guarded});
    this.dwvApp.addEventListener(type, guarded);
  }

  private failLoad(): void {
    this.dataLoaded = false;
    this.loadProgress = 0;
    this.loadError = 'Unable to display this DICOM document.';
  }

  ngOnDestroy(): void {
    this.destroyed = true;
    if (this.resizeHandler) window.removeEventListener('resize', this.resizeHandler);
    for (const {type, callback} of this.listeners) this.dwvApp.removeEventListener(type, callback);
    this.listeners = [];
    if (this.initialized) {
      this.dwvApp.abortAllLoads();
      this.dwvApp.reset();
    }
  }

  onChangeTool = (tool: string) => {
    if (!this.dwvApp || !this.dataLoaded || !this.canRunTool(tool)) return;
    this.selectedTool = tool;
    this.dwvApp.setTool(tool);
    if (tool === 'Draw') this.dwvApp.setToolFeatures({shapeName: this.tools.Draw.options[0]});
  };

  canRunTool = (tool: string) => {
    if (!this.dwvApp || !this.dataLoaded) return false;
    if (tool === 'Scroll') return this.dwvApp.canScroll();
    if (tool === 'WindowLevel') return this.dwvApp.canWindowLevel();
    return true;
  };

  toggleOrientation = () => {
    if (!this.dwvApp || !this.dataLoaded) return;
    this.orientation = this.orientation === 'coronal' ? 'sagittal' : this.orientation === 'sagittal' ? 'axial' : 'coronal';
    const config = new dwv.ViewConfig(this.layerGroupId);
    config.orientation = this.orientation;
    this.dwvApp.setDataViewConfigs({'*': [config]});
    for (const id of this.dwvApp.getDataIds()) this.dwvApp.render(id);
  };

  onReset = () => {
    if (this.dwvApp && this.dataLoaded) this.dwvApp.resetDisplay();
  };

  openTagsModal(template: TemplateRef<any>): void {
    if (!this.dataLoaded) return;
    this.refreshTags();
    this.modalService.open(template, {size: 'lg', ariaLabelledBy: 'modal-basic-title'});
  }

  refreshTags(): void {
    this.visibleMetaData = this.metaData.map((tag, i) => ({id: i + 1, ...tag})).slice(
      (this.tagsPage - 1) * this.tagsPageSize, this.tagsPage * this.tagsPageSize);
  }

  dataBase64toBlob(base64Data: string, contentType: string): Blob {
    const byteString = atob(base64Data);
    const bytes = new Uint8Array(byteString.length);
    for (let i = 0; i < byteString.length; i++) bytes[i] = byteString.charCodeAt(i);
    return new Blob([bytes], {type: contentType});
  }
}
