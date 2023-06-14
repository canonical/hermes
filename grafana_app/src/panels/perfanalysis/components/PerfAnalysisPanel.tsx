import React from 'react';
import { PanelProps } from "@grafana/data";
import { Options } from "../types";
import { TimeSeries, TooltipDisplayMode, TooltipPlugin, ZoomPlugin } from '@grafana/ui';

interface Props extends PanelProps<Options> { }

export const PerfAnalysisPanel = ({
  options,
  data,
  width,
  height,
  timeZone,
  timeRange,
  onChangeTimeRange,
}: Props) => {
  console.log(options)
  console.log(data)
  console.log(width)
  console.log(height)
  console.log(timeZone)
  console.log(timeRange)
  const dataAvailable = data?.series && data.series.length > 0;

  return (
    <div>
      {dataAvailable ? (
        <TimeSeries
          frames={data.series}
          timeRange={timeRange}
          timeZone={timeZone}
          width={width}
          height={height}
          legend={options.legend}
        >
          {(config, alignedDataFrame) => {
            return (
              <>
                <TooltipPlugin
                  config={config}
                  data={alignedDataFrame}
                  mode={TooltipDisplayMode.Multi}
                  timeZone={timeZone}
                />
                <ZoomPlugin config={config} onZoom={onChangeTimeRange} />
              </>
            );
          }}
        </TimeSeries>
      ) : (
        <div className="panel-empty">
          <p>No data to display.</p>
        </div>
      )}
    </div>
  );
}

