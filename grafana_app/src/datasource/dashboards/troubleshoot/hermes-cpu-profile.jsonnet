local perfAnalysisPanel = import '_perfanalysispanel.libsonnet';
local troubleshoot = import '_troubleshoot.libsonnet';

local routine = troubleshoot.getRoutineByID('hermes-cpu-profile');

troubleshoot.dashboard.new(routine)
.addPanel(
  perfAnalysisPanel.panel.new(
    datasource='$datasource',
    targets=routine.targets,
  ), gridPos={
    x: 0,
    y: 0,
    w: 24,
    h: 18,
  }
)
