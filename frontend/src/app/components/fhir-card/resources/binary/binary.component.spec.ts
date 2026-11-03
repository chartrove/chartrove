import { ComponentFixture, TestBed } from '@angular/core/testing';

import { BinaryComponent } from './binary.component';
import { HIGHLIGHT_OPTIONS } from 'ngx-highlightjs';
import {NgbCollapseModule} from '@ng-bootstrap/ng-bootstrap';
import {FastenApiService} from '../../../../services/fasten-api.service';
import {RouterTestingModule} from '@angular/router/testing';
import {HTTP_CLIENT_TOKEN} from '../../../../dependency-injection';
import { HttpClient, provideHttpClient, withInterceptorsFromDi, withXhr } from '@angular/common/http';
import { provideHttpClientTesting } from '@angular/common/http/testing';

describe('BinaryComponent', () => {
  let component: BinaryComponent;
  let fixture: ComponentFixture<BinaryComponent>;
  let mockedFastenApiService

  beforeEach(async () => {
    mockedFastenApiService = jasmine.createSpyObj('FastenApiService', ['getBinaryModel'])

    await TestBed.configureTestingModule({
    imports: [BinaryComponent, NgbCollapseModule, RouterTestingModule],
    providers: [
        {
            provide: HIGHLIGHT_OPTIONS,
            useValue: {
                coreLibraryLoader: () => import('highlight.js/lib/core'),
                languages: {
                    json: () => import('highlight.js/lib/languages/json'),
                    xml: () => import('highlight.js/lib/languages/xml'),
                },
            },
        },
        {
            provide: FastenApiService,
            useValue: mockedFastenApiService
        },
        {
            provide: HTTP_CLIENT_TOKEN,
            useClass: HttpClient,
        },
        provideHttpClient(withXhr(), withInterceptorsFromDi()),
        provideHttpClientTesting(),
    ]
})
    .compileComponents();

    fixture = TestBed.createComponent(BinaryComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  async function waitForHighlightedText(code: HTMLElement, expected: string, token: string) {
    const finished = () => code.textContent === expected && code.querySelector(token) !== null;
    if (finished()) return;
    await new Promise<void>((resolve, reject) => {
      const observer = new MutationObserver(() => {
        if (finished()) {
          observer.disconnect();
          clearTimeout(timer);
          resolve();
        }
      });
      const timer = setTimeout(() => {
        observer.disconnect();
        reject(new Error('Highlight DOM did not finish rendering: ' + code.textContent));
      }, 2000);
      observer.observe(code, {subtree: true, childList: true, characterData: true});
    });
  }

  async function renderSyntheticBinary(content_type: string, content: string) {
    const value = Object.freeze({content_type, content});
    fixture.componentRef.setInput('displayModel', value);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    expect(value.content).toBe(content);
    const code = fixture.nativeElement.querySelector('code') as HTMLElement;
    await waitForHighlightedText(code, content, content_type === 'application/xml' ? '.hljs-tag' : '.hljs-attr');
    return code;
  }

  it('highlights synthetic JSON binary content without activating embedded markup', async () => {
    const content = JSON.stringify({synthetic: true, value: 0, note: '<img src="x" onerror="throw 1">'});
    const code = await renderSyntheticBinary('application/json', content);
    expect(code.textContent).toBe(content);
    expect(code.querySelector('.hljs-attr')).not.toBeNull();
    expect(code.querySelector('img, script')).toBeNull();
  });

  it('highlights synthetic XML binary content and preserves its escaped text', async () => {
    const content = '<Observation><valueString value="Synthetic &lt;tag&gt;"/></Observation>';
    const code = await renderSyntheticBinary('application/xml', content);
    expect(code.textContent).toBe(content);
    expect(code.querySelector('.hljs-tag')).not.toBeNull();
    expect(code.querySelector('Observation, valueString')).toBeNull();
  });

  it('updates the binary renderer when JSON is replaced by synthetic XML', async () => {
    await renderSyntheticBinary('application/json', '{"synthetic":true}');
    const content = '<Synthetic value="0"/>';
    const code = await renderSyntheticBinary('application/xml', content);
    expect(code.textContent).toBe(content);
    expect(code.querySelector('.hljs-tag')).not.toBeNull();
  });

});
