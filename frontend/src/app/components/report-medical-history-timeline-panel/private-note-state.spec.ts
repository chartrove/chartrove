import {ComponentFixture, TestBed} from '@angular/core/testing';
import {MomentModule} from 'ngx-moment';
import {RouterTestingModule} from '@angular/router/testing';
import {NgbDropdownModule} from '@ng-bootstrap/ng-bootstrap';
import {ReportMedicalHistoryTimelinePanelComponent} from './report-medical-history-timeline-panel.component';

describe('Timeline private note availability', () => {
  let fixture: ComponentFixture<ReportMedicalHistoryTimelinePanelComponent>;
  let panel: HTMLElement;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [MomentModule, RouterTestingModule, NgbDropdownModule],
      declarations: [ReportMedicalHistoryTimelinePanelComponent],
    }).compileComponents();
    fixture = TestBed.createComponent(ReportMedicalHistoryTimelinePanelComponent);
    fixture.detectChanges();
    panel = fixture.nativeElement.querySelector('.timeline-comment-box');
  });

  it('does not invite writing a private note that cannot be persisted', () => {
    expect(panel.querySelector('input, textarea, [contenteditable="true"]')).toBeNull();
  });

  it('states note unavailability in visible text rather than a tooltip', () => {
    expect(panel.textContent).toContain('Private notes are not available yet.');
  });

  it('does not submit an unbound form when Enter is pressed', () => {
    expect(panel.querySelector('form')).toBeNull();
  });

  it('preserves the existing attachment entry action without submitting', () => {
    const openWizard = spyOn(fixture.componentInstance, 'openMedicalRecordWizard').and.stub();
    const button: HTMLButtonElement = panel.querySelector('button');
    expect(button.type).toBe('button');
    button.click();
    expect(openWizard).toHaveBeenCalledTimes(1);
  });
});
