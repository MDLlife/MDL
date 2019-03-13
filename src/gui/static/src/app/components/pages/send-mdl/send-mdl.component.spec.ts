import { async, ComponentFixture, TestBed } from '@angular/core/testing';

import { SendMDLComponent } from './send-mdl.component';

describe('SendMDLComponent', () => {
  let component: SendMDLComponent;
  let fixture: ComponentFixture<SendMDLComponent>;

  beforeEach(async(() => {
    TestBed.configureTestingModule({
      declarations: [ SendMDLComponent ],
    })
    .compileComponents();
  }));

  beforeEach(() => {
    fixture = TestBed.createComponent(SendMDLComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should be created', () => {
    expect(component).toBeTruthy();
  });
});
