local flamegraphPanel = import '_flamegraphpanel.libsonnet';
local timeSeriesPanel = import '_timeseriespanel.libsonnet';
local troubleshoot = import '_troubleshoot.libsonnet';

local routine = troubleshoot.getRoutineByID('hermes-memleak-profile');

troubleshoot.dashboard.new(routine)
.addPanel(
  timeSeriesPanel.panel.new(
    datasource='$datasource',
    target=routine.target,
  ), gridPos={
    x: 0,
    y: 0,
    w: 24,
    h: 8,
  }
)
.addPanel(
  flamegraphPanel.panel.new(
    datasource='$datasource',
    target=routine.target,
  ), gridPos={
    x: 0,
    y: 9,
    w: 24,
    h: 24,
  }
)
