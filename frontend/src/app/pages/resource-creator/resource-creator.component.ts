import {Component, OnInit, OnDestroy, ChangeDetectionStrategy} from '@angular/core';
import { NgbModal } from '@ng-bootstrap/ng-bootstrap';

import {Router} from '@angular/router';

@Component({
    selector: 'app-resource-creator',
    templateUrl: './resource-creator.component.html',
    styleUrls: ['./resource-creator.component.scss'],
    changeDetection: ChangeDetectionStrategy.Eager,
    standalone: false
})
export class ResourceCreatorComponent implements OnInit, OnDestroy {
  private destroyed = false;

  constructor( private modalService: NgbModal, private router: Router) { }

  async ngOnInit(): Promise<void> {
    const {MedicalRecordWizardComponent} = await import('../../components/medical-record-wizard/medical-record-wizard.component');
    if (this.destroyed) return;

    const modalRef = this.modalService.open(MedicalRecordWizardComponent, {
      size: 'xl',
    });

    //on close, we should redirect to the Medical history page.
    modalRef.result.then((result) => {
      this.router.navigate(['/medical-history']);
    }).catch((error) => {
      this.router.navigate(['/medical-history']);
    })
  }

  ngOnDestroy(): void {
    this.destroyed = true;
  }
}
