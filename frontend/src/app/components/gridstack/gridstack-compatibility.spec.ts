import { Component, Input } from '@angular/core';
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { GridStack, GridStackWidget } from 'gridstack';
import { DashboardComponent } from '../../pages/dashboard/dashboard.component';
import { GridstackComponent, NgGridStackWidget } from './gridstack.component';

@Component({
  selector: 'synthetic-grid-widget',
  template: '<span>{{ widgetConfig?.title_text }}</span>',
})
class SyntheticGridWidget {
  @Input() widgetConfig: any;
}

describe('Dashboard GridStack compatibility', () => {
  let fixture: ComponentFixture<GridstackComponent>;
  let previousTypes: typeof GridstackComponent.selectorToType;
  let previousAddRemove: typeof GridStack.addRemoveCB;
  let previousSave: typeof GridStack.saveCB;

  const widget: NgGridStackWidget = {
    id: 'synthetic-grid-a', x: 0, y: 0, w: 3, h: 2,
    type: 'synthetic-grid-widget',
    widgetConfig: {
      title_text: 'Synthetic layout widget',
      queries: [{ q: { from: 'Observation', select: ['valueQuantity.value'] } }],
      syntheticZero: 0, syntheticFalse: false,
    },
  };

  beforeEach(async () => {
    previousTypes = { ...GridstackComponent.selectorToType };
    previousAddRemove = GridStack.addRemoveCB;
    previousSave = GridStack.saveCB;
    await TestBed.configureTestingModule({ imports: [GridstackComponent, SyntheticGridWidget] }).compileComponents();
    GridstackComponent.addComponentToSelectorType([SyntheticGridWidget]);
  });

  afterEach(() => {
    fixture?.destroy();
    GridstackComponent.selectorToType = previousTypes;
    GridStack.addRemoveCB = previousAddRemove;
    GridStack.saveCB = previousSave;
  });

  function createGrid(width = 1200): GridStack {
    fixture = TestBed.createComponent(GridstackComponent);
    fixture.nativeElement.style.width = `${width}px`;
    // Consume the application's actual defaults, including responsive behavior.
    const dashboard = new DashboardComponent(null, null, null, null, null);
    fixture.componentInstance.options = { ...dashboard.gridOptions, column: 12, cellHeight: 40 };
    fixture.detectChanges();
    return fixture.componentInstance.grid;
  }

  function addSyntheticWidget(grid: GridStack): void {
    grid.addWidget(JSON.parse(JSON.stringify(widget)));
    fixture.detectChanges();
  }

  it('loads and round trips the registered widget type, geometry and zero-valued configuration', () => {
    const grid = createGrid();
    const raw = JSON.stringify(widget);
    addSyntheticWidget(grid);
    expect(fixture.nativeElement.textContent).toContain('Synthetic layout widget');
    const saved = grid.save(false) as NgGridStackWidget[];
    expect(saved.length).toBe(1);
    expect(saved[0]).toEqual(jasmine.objectContaining({
      id: 'synthetic-grid-a', x: 0, y: 0, w: 3, h: 2,
      type: 'synthetic-grid-widget', widgetConfig: widget.widgetConfig,
    }));
    grid.removeAll();
    grid.load(JSON.parse(JSON.stringify(saved)));
    fixture.detectChanges();
    expect(grid.getGridItems().length).toBe(1);
    expect((grid.save(false) as NgGridStackWidget[])[0].widgetConfig).toEqual(widget.widgetConfig);
    expect(JSON.stringify(widget)).toBe(raw);
  });

  it('keeps rendered geometry, layout callbacks and edit toggles working', async () => {
    const grid = createGrid();
    addSyntheticWidget(grid);
    const changed = jasmine.createSpy('changed');
    const removed = jasmine.createSpy('removed');
    fixture.componentInstance.changeCB.subscribe(changed);
    fixture.componentInstance.removedCB.subscribe(removed);
    const element = grid.getGridItems()[0];
    grid.update(element, { x: 4, y: 0, w: 5, h: 3 });
    expect((grid.save(false) as GridStackWidget[])[0])
      .toEqual(jasmine.objectContaining({ x: 4, y: 0, w: 5, h: 3 }));
    // Verify the shipped CSS as well as engine coordinates across the CSS-variable migration.
    for (let frame = 0; frame < 60; frame++) {
      const rect = element.getBoundingClientRect();
      if (Math.abs(rect.width - 500) < 0.5 && Math.abs(rect.height - 120) < 0.5) break;
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
    }
    const rect = element.getBoundingClientRect();
    expect(rect.width).toBeCloseTo(500, 0);
    expect(rect.height).toBeCloseTo(120, 0);
    expect(rect.left - grid.el.getBoundingClientRect().left).toBeCloseTo(400, 0);
    expect(changed).toHaveBeenCalled();
    grid.disable();
    expect(GridStack.getDD().isDraggable(element)).toBeFalse();
    expect(GridStack.getDD().isResizable(element)).toBeFalse();
    grid.enable();
    expect(GridStack.getDD().isDraggable(element)).toBeTrue();
    expect(GridStack.getDD().isResizable(element)).toBeTrue();
    grid.removeWidget(element);
    expect(grid.getGridItems().length).toBe(0);
    expect(removed).toHaveBeenCalled();
    expect(fixture.componentInstance.isEmpty).toBeTrue();
  });

  it('preserves the application narrow-screen one-column layout and restores twelve columns', async () => {
    const grid = createGrid(600);
    addSyntheticWidget(grid);
    async function expectColumns(columns: number): Promise<void> {
      for (let frame = 0; frame < 60 && grid.getColumn() !== columns; frame++) {
        await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
      }
      expect(grid.getColumn()).toBe(columns);
    }
    await expectColumns(1);
    fixture.nativeElement.style.width = '1200px';
    // GridStack animates host width. Dispatch the synthetic window resize only
    // after the requested geometry exists, rather than while it is still 600px.
    for (let frame = 0; frame < 60 && fixture.nativeElement.clientWidth !== 1200; frame++) {
      await new Promise<void>(resolve => requestAnimationFrame(() => resolve()));
    }
    expect(fixture.nativeElement.clientWidth).toBe(1200);
    window.dispatchEvent(new Event('resize'));
    await expectColumns(12);
    expect((grid.save(false) as GridStackWidget[])[0].w).toBe(3);
  });
});
