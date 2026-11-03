import { ComponentFixture, TestBed } from '@angular/core/testing';

import { DicomComponent } from './dicom.component';
import {BinaryModel} from 'src/lib/models/resources/binary-model';
import {syntheticExplicitDicom, syntheticRleDicom} from 'src/testing/synthetic-dicom-fixture';

describe('DicomComponent', () => {
  let component: DicomComponent;
  let fixture: ComponentFixture<DicomComponent>;
  const owned: ComponentFixture<DicomComponent>[] = [];

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ DicomComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(DicomComponent);
    owned.push(fixture);
    component = fixture.componentInstance;
  });

  afterEach(() => {
    // The old version has no component teardown. Dispose only these test-owned
    // viewers after assertions, so an inherited listener cannot leak across cases.
    for (const current of owned.splice(0)) {
      const app: any = current.componentInstance['dwvApp'];
      current.destroy();
      if (app) {
        window.removeEventListener('resize', app.onResize);
        app.abortAllLoads?.();
        app.reset();
      }
    }
  });

  async function render(current: ComponentFixture<DicomComponent>, data: string) {
    current.componentInstance.displayModel = new BinaryModel({contentType: 'application/dicom', data});
    current.detectChanges();
    await new Promise<void>((resolve, reject) => {
      const app: any = current.componentInstance['dwvApp'];
      const timer = setTimeout(() => reject(new Error('Synthetic DICOM render timed out')), 3000);
      let loaded = current.componentInstance.dataLoaded;
      let rendered = !!current.nativeElement.querySelector('canvas');
      function complete() { if (loaded && rendered) { clearTimeout(timer); resolve(); } }
      app.addEventListener('load', () => {loaded = true; complete();});
      app.addEventListener('renderend', () => {rendered = true; complete();});
      app.addEventListener('loaderror', () => {clearTimeout(timer); reject(new Error('Synthetic DICOM load failed'));});
      app.addEventListener('error', () => {clearTimeout(timer); reject(new Error('Synthetic DICOM load failed'));});
      complete();
    });
    current.detectChanges();
  }

  function imageOf(current: ComponentFixture<DicomComponent>) {
    const app: any = current.componentInstance['dwvApp'];
    const id = app.getDataIds ? app.getDataIds()[0] : 0;
    return app.getData ? app.getData(id).image : app.getImage(id);
  }

  for (const [label, data] of [['explicit little-endian', syntheticExplicitDicom], ['lossless RLE', syntheticRleDicom]]) {
    it(`renders actual ${label} pixels and preserves the original Binary`, async () => {
      await render(fixture, data);
      const image = imageOf(fixture);
      expect(image.getGeometry().getSize().get(0)).toBe(16);
      expect(image.getGeometry().getSize().get(1)).toBe(16);
      expect(image.getGeometry().getSpacing().get(0)).toBe(0.5);
      expect(Array.from({length: 256}, (_, i) => image.getValueAtOffset(i)))
        .toEqual(Array.from({length: 256}, (_, i) => [0, 128, 512, 1023][i % 4]));
      expect(fixture.nativeElement.querySelector('canvas')).not.toBeNull();
      expect(component.displayModel.data).toBe(data);
      component.refreshTags();
      expect(component.visibleMetaData.length).toBeGreaterThan(0);
      expect(JSON.stringify(component['metaData'])).toContain('SYNTHETIC^CHARTROVE');
      component.onChangeTool('WindowLevel');
      expect(component.selectedTool).toBe('WindowLevel');
      component.onChangeTool('Draw');
      expect(component.selectedTool).toBe('Draw');
      expect(() => component.onReset()).not.toThrow();
      expect(component.displayModel.data).toBe(data);
    });
  }

  it('keeps simultaneous viewers and their canvas containers separate', async () => {
    await render(fixture, syntheticExplicitDicom);
    const second = TestBed.createComponent(DicomComponent);
    owned.push(second);
    await render(second, syntheticRleDicom);
    const firstCanvas = fixture.nativeElement.querySelector('canvas');
    const secondCanvas = second.nativeElement.querySelector('canvas');
    expect(firstCanvas).not.toBeNull();
    expect(secondCanvas).not.toBeNull();
    expect(firstCanvas).not.toBe(secondCanvas);
    expect(component.displayModel.data).toBe(syntheticExplicitDicom);
    expect(second.componentInstance.displayModel.data).toBe(syntheticRleDicom);
  });

  it('releases its resize listener when the component is destroyed', async () => {
    await render(fixture, syntheticExplicitDicom);
    const resize = component['dwvApp'].onResize;
    const remove = spyOn(window, 'removeEventListener').and.callThrough();
    fixture.destroy();
    expect(remove).toHaveBeenCalledWith('resize', resize);
  });

  it('rerenders orientation controls without replacing the original Binary', async () => {
    await render(fixture, syntheticExplicitDicom);
    const button = fixture.nativeElement.querySelector('button[ngbtooltip="toggle orientation"]') as HTMLButtonElement;
    expect(button.disabled).toBeFalse();
    for (let i = 0; i < 3; i++) {
      expect(() => button.click()).not.toThrow();
      fixture.detectChanges();
      expect(fixture.nativeElement.querySelector('canvas')).not.toBeNull();
      expect(imageOf(fixture).getValueAtOffset(0)).toBe(0);
      expect(component.displayModel.data).toBe(syntheticExplicitDicom);
    }
  });

  it('shows a safe error and disables tools for a malformed encoded document', async () => {
    const alert = spyOn(window, 'alert');
    component.displayModel = new BinaryModel({});
    component.displayModel.data = '%%%synthetic-invalid-base64';
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    expect(component.dataLoaded).toBeFalse();
    expect(fixture.nativeElement.querySelector('[role="alert"]').textContent)
      .toContain('Unable to display this DICOM document.');
    expect(fixture.nativeElement.textContent).not.toContain('synthetic-invalid');
    expect(fixture.nativeElement.querySelector('button[ngbtooltip="change brightness"]').disabled).toBeTrue();
    expect(component.displayModel.data).toBe('%%%synthetic-invalid-base64');
    expect(alert).not.toHaveBeenCalled();
  });

  it('should create', () => {
    fixture.detectChanges();
    expect(component).toBeTruthy();
  });
});
