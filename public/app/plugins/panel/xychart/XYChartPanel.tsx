import { css } from '@emotion/css';
import React, { useState, useEffect, useCallback } from 'react';
import { usePrevious } from 'react-use';

import {
  DisplayProcessor,
  DisplayValue,
  fieldReducers,
  PanelProps,
  reduceField,
  ReducerID,
  getDisplayProcessor,
} from '@grafana/data';
import { alpha } from '@grafana/data/src/themes/colorManipulator';
import { config } from '@grafana/runtime';
import {
  TooltipDisplayMode,
  TooltipPlugin2,
  UPlotChart,
  UPlotConfigBuilder,
  VizLayout,
  VizLegend,
  VizLegendItem,
} from '@grafana/ui';
import { TooltipHoverMode } from '@grafana/ui/src/components/uPlot/plugins/TooltipPlugin2';
import { FacetedData } from '@grafana/ui/src/components/uPlot/types';

import { XYChartTooltip } from './XYChartTooltip';
import { Options, SeriesMapping } from './panelcfg.gen';
import { prepData, prepScatter, ScatterPanelInfo } from './scatter';
import { ScatterSeries } from './types';

type Props = PanelProps<Options>;

export const XYChartPanel = (props: Props) => {
  const [error, setError] = useState<string | undefined>();
  const [series, setSeries] = useState<ScatterSeries[]>([]);
  const [builder, setBuilder] = useState<UPlotConfigBuilder | undefined>();
  const [facets, setFacets] = useState<FacetedData | undefined>();

  const oldOptions = usePrevious(props.options);
  const oldData = usePrevious(props.data);

  const initSeries = useCallback(() => {
    const getData = () => props.data.series;
    const info: ScatterPanelInfo = prepScatter(props.options, getData, config.theme2);

    if (info.error) {
      setError(info.error);
    } else if (info.series.length && props.data.series) {
      setBuilder(info.builder);
      setSeries(info.series);
      setFacets(() => prepData(info, props.data.series));
      setError(undefined);
    }
  }, [props.data.series, props.options]);

  const initFacets = useCallback(() => {
    setFacets(() => prepData({ error, series }, props.data.series));
  }, [props.data.series, error, series]);

  useEffect(() => {
    if (oldOptions !== props.options || oldData?.structureRev !== props.data.structureRev) {
      initSeries();
    } else if (oldData?.series !== props.data.series) {
      initFacets();
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props]);

  const renderLegend = () => {
    const items: VizLegendItem[] = [];
    const defaultFormatter = (v: any) => (v == null ? '-' : v.toFixed(1));
    const theme = config.theme2;

    for (let si = 0; si < series.length; si++) {
      const s = series[si];
      const frame = s.frame(props.data.series);
      if (frame) {
        for (const item of s.legend()) {
          item.getDisplayValues = () => {
            const calcs = props.options.legend.calcs;

            if (!calcs?.length) {
              return [];
            }

            const field = s.y(frame);

            const fmt = field.display ?? defaultFormatter;
            let countFormatter: DisplayProcessor | null = null;

            const fieldCalcs = reduceField({
              field,
              reducers: calcs,
            });

            return calcs.map<DisplayValue>((reducerId) => {
              const fieldReducer = fieldReducers.get(reducerId);
              let formatter = fmt;

              if (fieldReducer.id === ReducerID.diffperc) {
                formatter = getDisplayProcessor({
                  field: {
                    ...field,
                    config: {
                      ...field.config,
                      unit: 'percent',
                    },
                  },
                  theme,
                });
              }

              if (
                fieldReducer.id === ReducerID.count ||
                fieldReducer.id === ReducerID.changeCount ||
                fieldReducer.id === ReducerID.distinctCount
              ) {
                if (!countFormatter) {
                  countFormatter = getDisplayProcessor({
                    field: {
                      ...field,
                      config: {
                        ...field.config,
                        unit: 'none',
                      },
                    },
                    theme,
                  });
                }
                formatter = countFormatter;
              }

              return {
                ...formatter(fieldCalcs[reducerId]),
                title: fieldReducer.name,
                description: fieldReducer.description,
              };
            });
          };

          item.disabled = !(s.show ?? true);

          if (props.options.seriesMapping === SeriesMapping.Manual) {
            item.label = props.options.series?.[si]?.name ?? `Series ${si + 1}`;
          }

          item.color = alpha(s.lineColor(frame) as string, 1);

          items.push(item);
        }
      }
    }

    if (!props.options.legend.showLegend) {
      return null;
    }

    const legendStyle = {
      flexStart: css({
        div: {
          justifyContent: 'flex-start',
        },
      }),
    };

    return (
      <VizLayout.Legend placement={props.options.legend.placement} width={props.options.legend.width}>
        <VizLegend
          className={legendStyle.flexStart}
          placement={props.options.legend.placement}
          items={items}
          displayMode={props.options.legend.displayMode}
        />
      </VizLayout.Legend>
    );
  };

  if (error || !builder || !facets) {
    return (
      <div className="panel-empty">
        <p>{error}</p>
      </div>
    );
  }

  return (
    <>
      <VizLayout width={props.width} height={props.height} legend={renderLegend()}>
        {(vizWidth: number, vizHeight: number) => (
          <UPlotChart config={builder} data={facets} width={vizWidth} height={vizHeight}>
            {props.options.tooltip.mode !== TooltipDisplayMode.None && (
              <TooltipPlugin2
                config={builder}
                hoverMode={TooltipHoverMode.xyOne}
                render={(u, dataIdxs, seriesIdx, isPinned, dismiss) => {
                  return (
                    <XYChartTooltip
                      data={props.data.series}
                      dataIdxs={dataIdxs}
                      allSeries={series}
                      dismiss={dismiss}
                      isPinned={isPinned}
                      options={props.options}
                      seriesIdx={seriesIdx}
                    />
                  );
                }}
                maxWidth={props.options.tooltip.maxWidth}
              />
            )}
          </UPlotChart>
        )}
      </VizLayout>
    </>
  );
};