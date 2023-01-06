// THIS FILE IS GENERATED. EDITING IS FUTILE.
//
// Generated by:
//     public/app/plugins/gen.go
// Using jennies:
//     TSTypesJenny
//     PluginTSTypesJenny
//
// Run 'make gen-cue' from repository root to regenerate.

import * as ui from '@grafana/schema';

export const PanelModelVersion = Object.freeze([0, 0]);

/**
 * Select the pie chart display style.
 */
export enum PieChartType {
  Donut = 'donut',
  Pie = 'pie',
}

/**
 * Select labels to display on the pie chart.
 *  - Name - The series or field name.
 *  - Percent - The percentage of the whole.
 *  - Value - The raw numerical value.
 */
export enum PieChartLabels {
  Name = 'name',
  Percent = 'percent',
  Value = 'value',
}

/**
 * Select values to display in the legend.
 *  - Percent: The percentage of the whole.
 *  - Value: The raw numerical value.
 */
export enum PieChartLegendValues {
  Percent = 'percent',
  Value = 'value',
}

export interface PieChartLegendOptions extends ui.VizLegendOptions {
  values: Array<PieChartLegendValues>;
}

export const defaultPieChartLegendOptions: Partial<PieChartLegendOptions> = {
  values: [],
};

export interface PanelOptions extends ui.OptionsWithTooltip, ui.SingleStatBaseOptions {
  displayLabels: Array<PieChartLabels>;
  legend: PieChartLegendOptions;
  pieType: PieChartType;
}

export const defaultPanelOptions: Partial<PanelOptions> = {
  displayLabels: [],
};

export interface PanelFieldConfig extends ui.HideableFieldConfig {}
