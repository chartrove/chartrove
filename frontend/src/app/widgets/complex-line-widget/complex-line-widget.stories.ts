import {applicationConfig, Meta, StoryObj} from '@storybook/angular';
import {of} from 'rxjs';
import {ComplexLineWidgetComponent} from './complex-line-widget.component';
import {FastenApiService} from '../../services/fasten-api.service';

const meta: Meta<ComplexLineWidgetComponent> = {
  title: 'Widget/ComplexLineWidget',
  component: ComplexLineWidgetComponent,
  decorators: [applicationConfig({providers: [{
    provide: FastenApiService,
    useValue: {queryResources: () => of({success: true, data: [0, 2.5].map((value, index) => ({
      resource_raw: {resourceType: 'Observation', id: `synthetic-${index}`,
        issued: `2000-01-0${index + 1}T00:00:00Z`, valueQuantity: {value, unit: 'kg'}},
    }))})},
  }]})],
  tags: ['autodocs'],
  argTypes: {widgetConfig: {control: 'object'}},
};

export default meta;
type Story = StoryObj<ComplexLineWidgetComponent>;

export const Example: Story = {
  args: {
    widgetConfig: {
      item_type: 'complex-line-widget', title_text: 'Synthetic readings',
      description_text: 'Disposable example records, including zero.',
      width: 8, height: 4,
      queries: [{q: {from: 'Observation', select: ['issued as label', 'valueQuantity.value as data'],
        where: {code: 'synthetic-quantity'}}, dataset_options: {label: 'Synthetic kg'}}],
      parsing: {xAxisKey: 'label', yAxisKey: 'data'},
    },
  },
};
