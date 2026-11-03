import { ComponentFixture, TestBed } from '@angular/core/testing';
import {Component, ChangeDetectionStrategy} from '@angular/core';

import { RtfComponent } from './rtf.component';
import { BinaryModel } from 'src/lib/models/resources/binary-model';
import { RTFJS } from 'rtf.js';

function documentModel(text: string): BinaryModel {
  const model = new BinaryModel({});
  model.content = `{\\rtf1\\ansi ${text}}`;
  return model;
}

@Component({
  changeDetection: ChangeDetectionStrategy.Eager,
  imports: [RtfComponent],
  template: `@if (showFirst) { <fhir-rtf [displayModel]="first" /> }
    <fhir-rtf [displayModel]="second" />`,
})
class RtfPairHost {
  showFirst = true;
  first = documentModel('SYNTHETIC FIRST');
  second = documentModel('SYNTHETIC SECOND');
}

describe('RtfComponent', () => {
  let component: RtfComponent;
  let fixture: ComponentFixture<RtfComponent>;

  const rtf: string = "{\\rtf1\\ansi\\deff0 {\\fonttbl {\\f0 Times New Roman;}} \\f0\\fs60 Hello, World! }";

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ RtfComponent, RtfPairHost ]
    })
    .compileComponents();

    RTFJS.loggingEnabled(false);

    fixture = TestBed.createComponent(RtfComponent);
    component = fixture.componentInstance;
    component.displayModel = new BinaryModel({});
    component.displayModel.content = rtf;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  function waitForText(root: HTMLElement, text: string): Promise<void> {
    if (root.textContent.includes(text)) return Promise.resolve();
    return new Promise((resolve, reject) => {
      const observer = new MutationObserver(() => {
        if (!root.textContent.includes(text)) return;
        observer.disconnect();
        clearTimeout(timer);
        resolve();
      });
      const timer = setTimeout(() => {
        observer.disconnect();
        reject(new Error(`Synthetic RTF text was not rendered: ${text}`));
      }, 3000);
      observer.observe(root, {childList: true, subtree: true, characterData: true});
    });
  }

  it('renders the actual document text without changing its source', async () => {
    await waitForText(fixture.nativeElement, 'Hello, World!');
    expect(component.displayModel.content).toBe(rtf);
  });

  it('keeps simultaneous documents in their own component containers', async () => {
    await waitForText(fixture.nativeElement, 'Hello, World!');
    fixture.destroy();
    const pair = TestBed.createComponent(RtfPairHost);
    const firstSource = pair.componentInstance.first.content;
    const secondSource = pair.componentInstance.second.content;
    pair.detectChanges();
    const [first, second] = Array.from(pair.nativeElement.querySelectorAll('fhir-rtf')) as HTMLElement[];
    try {
      const results = await Promise.allSettled([
        waitForText(first, 'SYNTHETIC FIRST'),
        waitForText(second, 'SYNTHETIC SECOND'),
      ]);
      for (const result of results) if (result.status === 'rejected') throw result.reason;
      expect(first.textContent).not.toContain('SYNTHETIC SECOND');
      expect(second.textContent).not.toContain('SYNTHETIC FIRST');
      expect(pair.componentInstance.first.content).toBe(firstSource);
      expect(pair.componentInstance.second.content).toBe(secondSource);
    } finally {
      pair.destroy();
    }
  });

  it('does not append a destroyed document into another record', async () => {
    await waitForText(fixture.nativeElement, 'Hello, World!');
    fixture.destroy();
    const pair = TestBed.createComponent(RtfPairHost);
    pair.componentInstance.first = documentModel('SYNTHETIC STALE');
    pair.componentInstance.second = documentModel('SYNTHETIC ACTIVE');
    pair.detectChanges();
    const detachedRoot = pair.nativeElement.querySelector('fhir-rtf') as HTMLElement;
    pair.componentInstance.showFirst = false;
    pair.detectChanges();
    expect(pair.nativeElement.querySelectorAll('fhir-rtf').length).toBe(1);
    expect(detachedRoot.isConnected).toBeFalse();
    const active = pair.nativeElement.querySelector('fhir-rtf') as HTMLElement;
    try {
      await waitForText(active, 'SYNTHETIC ACTIVE');
      await pair.whenStable();
      expect(active.textContent).not.toContain('SYNTHETIC STALE');
      expect(detachedRoot.textContent).not.toContain('SYNTHETIC STALE');
    } finally {
      pair.destroy();
    }
  });
});
