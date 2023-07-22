{
  panel: {
    new(
      datasource,
      target
    ):: {
      type: 'hermes-flamegraph-panel',
      datasource: datasource,
      targets: [target],
      options: {
        group: target.group,
        routine: target.routine,
        ds_id: 'grafana-hermes-datasource',
      },
    }
  }
}
