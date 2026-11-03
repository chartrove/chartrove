import {Component, Input, OnInit, OnDestroy, ViewChild, ElementRef, ChangeDetectionStrategy} from '@angular/core';

import {BinaryModel} from '../../../../../lib/models/resources/binary-model';

@Component({
    imports: [],
    selector: 'fhir-rtf',
    templateUrl: './rtf.component.html',
    changeDetection: ChangeDetectionStrategy.Eager,
    styleUrls: ['./rtf.component.scss']
})
export class RtfComponent implements OnInit, OnDestroy {
  @Input() displayModel: BinaryModel
  @ViewChild('rtfContent', {static: true}) private rtfContent: ElementRef<HTMLElement>;
  private destroyed = false;

  constructor() { }

  async ngOnInit(): Promise<void> {
    const source = this.displayModel.content;
    const {RTFJS} = await import('rtf.js');
    if (this.destroyed) return;
    const doc = new RTFJS.Document(this.stringToArrayBuffer(source), null);
    try {
      const htmlElements = await doc.render();
      if (this.destroyed) return;
      for (const element of htmlElements) {
        this.rtfContent.nativeElement.appendChild(element);
      }
    } catch (e) {
      if (this.destroyed) return;
      if (e instanceof RTFJS.Error) {
        console.error("Error: " + e.message);
        this.rtfContent.nativeElement.textContent = e.message;
      }
      else {
        throw e;
      }
    }
  }

  ngOnDestroy(): void {
    this.destroyed = true;
  }

  private stringToArrayBuffer(string): ArrayBuffer {
    const buffer = new ArrayBuffer(string.length);
    const bufferView = new Uint8Array(buffer);
    for (let i = 0; i < string.length; i++) {
      bufferView[i] = string.charCodeAt(i);
    }
    return buffer;
  }

}
