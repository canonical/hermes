local grafana = import 'grafonnet/grafana.libsonnet';
local troubleshoot = import '_troubleshoot.libsonnet';

{
  routines: [
    {
      id: 'hermes-cpu-profile',
      title: 'CPU Profile',
      target: {refId: 'A', group: 'cpu', routine: 'cpu_profile'},
    },
    {
      id: 'hermes-memleak-profile',
      title: 'Memleak Profile',
      target: {refId: 'A', group: 'memory', routine: 'memleak_profile'},
    },
  ],

  getRoutineByID(id)::
    local result = std.filter(function(x) x.id == id, self.routines);
    if std.length(result) == 0 then {} else result[0],

  dashboard: {
    new(routine)::
      grafana.dashboard.new(
        routine.title,
        time_from='now-15m',
        time_to='now',
        refresh='10s',
        timepicker=grafana.timepicker.new(
          refresh_intervals=['1s', '2s', '5s', '10s'],
        ),
      )
      .addTemplate(
        grafana.template.datasource(
          'datasource',
          'grafana-hermes-datasource',
          'Grafana Hermes Datasource',
        )
      ),
  }
}
