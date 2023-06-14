import React from 'react';
import { DataSourcePluginOptionsEditorProps } from '@grafana/data';
import { DataSourceHttpSettings } from '@grafana/ui';
import { DataSourceOptions } from '../types';

type Props = DataSourcePluginOptionsEditorProps<DataSourceOptions>;

export const ConfigEditor = React.memo((props: Props) => {
  return (
    <div className="gf-form-group">
      <div className="gf-form">
        <DataSourceHttpSettings
          defaultUrl="http://127.0.0.1:8080"
          dataSourceConfig={props.options}
          onChange={props.onOptionsChange}
        />
      </div>
    </div>

  );
});
ConfigEditor.displayName = 'ConfigEditor';
