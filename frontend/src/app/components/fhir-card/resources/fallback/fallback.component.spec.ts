import { ComponentFixture, TestBed } from '@angular/core/testing';

import { FallbackComponent } from './fallback.component';
import { HIGHLIGHT_OPTIONS } from 'ngx-highlightjs';

describe('FallbackComponent', () => {
  let component: FallbackComponent;
  let fixture: ComponentFixture<FallbackComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ FallbackComponent ],
      providers: [
        {
          provide: HIGHLIGHT_OPTIONS,
          useValue: {
            coreLibraryLoader: () => import('highlight.js/lib/core'),
            languages: {
              json: () => import('highlight.js/lib/languages/json')
            },
          }
        },
      ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(FallbackComponent);
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

  async function renderSyntheticModel(value: unknown) {
    fixture.componentRef.setInput('displayModel', value);
    fixture.detectChanges();
    await fixture.whenStable();
    fixture.detectChanges();
    const code = fixture.nativeElement.querySelector('code') as HTMLElement;
    await waitForHighlightedText(code, JSON.stringify(value, null, 2), '.hljs-attr');
    return code;
  }

  it('highlights synthetic JSON while preserving zero, false and Unicode text', async () => {
    const value = Object.freeze({resourceType: 'SyntheticResource', valueInteger: 0,
      active: false, note: 'Synthetic café مثال'});
    const original = JSON.stringify(value);
    const code = await renderSyntheticModel(value);
    expect(code.textContent).toBe(JSON.stringify(value, null, 2));
    expect(code.querySelector('.hljs-attr')).not.toBeNull();
    expect(code.querySelector('.hljs-literal')).not.toBeNull();
    expect(JSON.stringify(value)).toBe(original);
  });

  it('refreshes the highlighted JSON when the selected synthetic resource changes', async () => {
    await renderSyntheticModel({resourceType: 'SyntheticResource', valueInteger: 0});
    const next = {resourceType: 'SyntheticResource', valueInteger: 17, active: false};
    const code = await renderSyntheticModel(next);
    expect(code.textContent).toBe(JSON.stringify(next, null, 2));
    expect(code.querySelector('.hljs-number')?.textContent).toBe('17');
  });

  it('keeps markup in a synthetic record as literal text without creating elements', async () => {
    const value = {resourceType: 'SyntheticResource', note: '<img src="x" onerror="throw 1"><script>throw 2</script>'};
    const code = await renderSyntheticModel(value);
    expect(code.textContent).toBe(JSON.stringify(value, null, 2));
    expect(code.querySelector('img, script')).toBeNull();
    expect(code.querySelector('.hljs-string')).not.toBeNull();
  });

});
