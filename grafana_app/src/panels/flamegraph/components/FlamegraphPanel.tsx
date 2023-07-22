import React, { useState } from 'react';
import { PanelProps } from "@grafana/data";
import { Select } from '@grafana/ui';
import { getDataSourceSrv } from '@grafana/runtime';
import { FlamegraphRenderer, Box } from "@pyroscope/flamegraph";
import { Options } from "../types";

import '@pyroscope/flamegraph/dist/index.css';

interface Props extends PanelProps<Options> { }

export const FlamegraphPanel = ({
  options,
  data,
  width,
  height,
  timeZone,
  timeRange,
  onChangeTimeRange,
}: Props) => {
  const [selectedOption, setSelectedOption] = useState<string>('');
  const [profile, setProfile] = useState<object>({});
  const dataAvailable = data?.series && data.series.length > 0;
  const profileAvailable = Object.keys(profile).length !== 0;
  const dropdownOptions = []
  const dataSourceSrv: any = getDataSourceSrv();
  let datasource: any = null;
  const timestampToStr = (timestamp: number) => {
    const date = new Date(timestamp)
    return date.getHours() + ":" + date.getMinutes() + ":" + date.getSeconds() + ", " + date.toDateString()
  }
  const handleOptionChange = async (selected: string) => {
    setSelectedOption(selected)
    const _profile = await datasource.getResource([options.group, options.routine, selected].join('/'))
    setProfile(_profile)
  }

  Object.keys(dataSourceSrv.datasources).forEach((key: string) => {
    if (dataSourceSrv.datasources[key].type === options.ds_id) {
      datasource = dataSourceSrv.datasources[key]
    }
  });

  if (dataAvailable) {
    let len = data.series[0].length
    for (let i = 0; i < len; ++i) {
      let timestamp = data.series[0].fields[0].values.get(i)
      let threshold = data.series[0].fields[1].values.get(i)
      let usage = data.series[1].fields[1].values.get(i)
      if (usage >= threshold) {
        dropdownOptions.push(
          { label: timestampToStr(timestamp), value: timestamp }
        )
      }
    }
  }

  return (
    <div>
      {dataAvailable ? (
        <>
          <Select
            options={dropdownOptions}
            value={selectedOption}
            onChange={(option) => handleOptionChange(option.value as string)}
          />
          {
            profileAvailable ? (
              <Box>
                <FlamegraphRenderer
                  profile={profile}
                  onlyDisplay="flamegraph"
                  showToolbar={false}
                />
              </Box>
            ) : (
              <div className="panel-empty">
                <p>No profile to display.</p>
              </div >
            )
          }
        </>
      ) : (
        <div className="panel-empty">
          <p>No data to display.</p>
        </div >
      )
      }
    </div >
  )
}

