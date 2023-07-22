{
  panel: {
    new(
      datasource,
      target
    ):: {
      type: 'hermes-time-series-panel',
      datasource: datasource,
      targets: [target],
    }
  }
}
