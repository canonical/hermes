# Introduction
`hermes` is a time-series profiling tool for analyzing system performance issues from different dimensions.

# Building
`hermes` shall be built on Ubuntu Focal or higher versions because of package dependency.
```
git clone https://github.com/yukariatlas/hermes.git
cd hermes
make
# install binaries and config
make install_bin
# install UI
make install_ui
```

# Running
`hermes` is composed of three different components. Please follow the steps to get a performance analysis result.
### Collecting metrics
`collector` is a binary to collect system performance metrics. The metrics will store under /var/log/collector folder.
### Parsing metrics
`parser` is a tool to parse the collected metrics into a specific format for UI to show graphs. The parsed data will store under $HOME/view folder.
### UI
`webserver` is a binary to show data under $HOME/view to UI. It can be run on localhost and use `http://127.0.0.1:8080/` to see the performance analysis result.
