import { DataQuery, DataSourceJsonData } from '@grafana/data';

export interface Query extends DataQuery { }

export interface DataSourceOptions extends DataSourceJsonData { }
