import { ComponentFixture, TestBed } from '@angular/core/testing';

import { InterpreteComponent } from './interprete.component';

describe('InterpreteComponent', () => {
  let component: InterpreteComponent;
  let fixture: ComponentFixture<InterpreteComponent>;

  beforeEach(async () => {
    await TestBed.configureTestingModule({
      declarations: [ InterpreteComponent ]
    })
    .compileComponents();

    fixture = TestBed.createComponent(InterpreteComponent);
    component = fixture.componentInstance;
    fixture.detectChanges();
  });

  it('should create', () => {
    expect(component).toBeTruthy();
  });
});
