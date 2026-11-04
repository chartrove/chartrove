import { ComponentFixture, TestBed } from '@angular/core/testing';

import { BinaryTextComponent } from './binary-text.component';
import {BinaryModel} from '../../../../../lib/models/resources/binary-model';

describe('BinaryTextComponent', () => {
  let component: BinaryTextComponent;
  let fixture: ComponentFixture<BinaryTextComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      imports: [ BinaryTextComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(BinaryTextComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });

  it('renders a reviewed UTF-8 transcript exactly as text while preserving binary bytes', () => {
    const text='Synthetic café — العربية <em>reviewed</em>';
    const encoded=btoa(String.fromCharCode(...new TextEncoder().encode(text)));
    component.displayModel=new BinaryModel({contentType:'text/plain',data:encoded});fixture.detectChanges();
    expect(component.displayModel.content).toBe(text);
    expect(fixture.nativeElement.textContent).toContain(text);expect(fixture.nativeElement.querySelector('em')).toBeNull();
    expect(new BinaryModel({contentType:'application/pdf',data:encoded}).content).toBe(atob(encoded));
    expect(new BinaryModel({contentType:'text/plain',data:btoa('\xff')}).content).toBe('\xff');
  });
});
