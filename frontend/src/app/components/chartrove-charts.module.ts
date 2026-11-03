import { NgModule } from '@angular/core';
import { BaseChartDirective, provideCharts, withDefaultRegisterables } from 'ng2-charts';

// Keep registration available to standalone widgets, clinical cards and Storybook.
@NgModule({
  imports: [BaseChartDirective],
  exports: [BaseChartDirective],
  providers: [provideCharts(withDefaultRegisterables())],
})
export class ChartroveChartsModule {}
