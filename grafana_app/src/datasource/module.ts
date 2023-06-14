import { DataSourcePlugin } from '@grafana/data';
import { DataSource } from './datasource';
import { ConfigEditor } from './components/ConfigEditor';

export const plugin = new DataSourcePlugin(DataSource)
  .setConfigEditor(ConfigEditor)
