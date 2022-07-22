import { act, render, screen } from '@testing-library/react';
import React, { FC } from 'react';
import { Provider } from 'react-redux';
import configureMockStore from 'redux-mock-store';
import { ReplaySubject } from 'rxjs';

import {
  dateTime,
  EventBusSrv,
  getDefaultTimeRange,
  LoadingState,
  PanelData,
  PanelPlugin,
  PanelProps,
  TimeRange,
} from '@grafana/data';

import { PanelQueryRunner } from '../../../query/state/PanelQueryRunner';
import { setTimeSrv, TimeSrv } from '../../services/TimeSrv';
import { DashboardModel, PanelModel } from '../../state';

import { PanelEditorTableView } from './PanelEditorTableView';

jest.mock('app/core/profiler', () => ({
  profiler: {
    renderingCompleted: jest.fn(),
  },
}));

jest.mock('app/features/panel/components/PanelRenderer', () => ({
  PanelRenderer: jest.fn(() => <div>PanelRenderer</div>),
}));

jest.mock('app/features/dashboard/utils/panel', () => ({
  applyPanelTimeOverrides: jest.fn(),
}));

function setupTestContext(options) {
  const mockStore = configureMockStore();
  const store = mockStore({ dashboard: { panels: [] } });
  const subject: ReplaySubject<PanelData> = new ReplaySubject<PanelData>();
  const panelQueryRunner = {
    getData: () => subject,
    run: () => {
      subject.next({ state: LoadingState.Done, series: [], timeRange: getDefaultTimeRange() });
    },
  } as unknown as PanelQueryRunner;
  const timeSrv = {
    timeRange: jest.fn(),
  } as unknown as TimeSrv;
  setTimeSrv(timeSrv);

  const defaults = {
    panel: new PanelModel({
      id: 123,
      hasTitle: jest.fn(),
      replaceVariables: jest.fn(),
      events: new EventBusSrv(),
      getQueryRunner: () => panelQueryRunner,
      getOptions: jest.fn(),
      getDisplayTitle: jest.fn(),
      timeFrom: jest.fn(),
      timeTo: jest.fn(),
    }),
    dashboard: {
      panelInitialized: jest.fn(),
      getTimezone: () => 'browser',
      events: new EventBusSrv(),
      meta: {
        isPublic: false,
      },
    } as unknown as DashboardModel,
    plugin: {
      meta: { skipDataQuery: false },
      panel: TestPanelComponent,
    } as unknown as PanelPlugin,
    isViewing: false,
    isEditing: true,
    isInView: false,
    width: 100,
    height: 100,
    onInstanceStateChange: () => {},
  };

  const props = { ...defaults, ...options };
  const { rerender } = render(
    <Provider store={store}>
      <PanelEditorTableView {...props} />
    </Provider>
  );

  return { rerender, props, subject, store };
}

describe('PanelEditorTableView', () => {
  it('should render', () => {
    const { rerender, props, subject, store } = setupTestContext({});

    const timeRangeUpdated = {
      from: dateTime([2019, 1, 11, 12, 0]),
      to: dateTime([2019, 1, 11, 18, 0]),
      raw: {
        from: 'now-6h',
        to: 'now',
      },
    } as unknown as TimeRange;

    act(() => {
      subject.next({ state: LoadingState.Loading, series: [], timeRange: getDefaultTimeRange() });
      subject.next({ state: LoadingState.Done, series: [], timeRange: getDefaultTimeRange() });
    });

    const newProps = { ...props, isInView: true };
    rerender(
      <Provider store={store}>
        <PanelEditorTableView {...newProps} />
      </Provider>
    );

    expect(screen.getByText(/PanelRenderer/i)).toBeInTheDocument();

    // triggering refresh events should update time range
    act(() => {
      subject.next({ state: LoadingState.Loading, series: [], timeRange: timeRangeUpdated });
      subject.next({ state: LoadingState.Done, series: [], timeRange: getDefaultTimeRange() });
    });

    expect(props.panel.timeFrom).toHaveBeenCalled();
    expect(props.panel.timeTo).toHaveBeenCalled();
  });
});

const TestPanelComponent: FC<PanelProps> = () => <div>Plugin Panel to Render</div>;
