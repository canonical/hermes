# Introduction
Hermes is a versatile time-series profiling tool for comprehensive system performance analysis. It runs periodic or oneshot jobs defined in the configurations to monitor performance metrics. If system loading stays below defined thresholds, metric collection is skipped to ensure minimal impact on system performance. This allows the profiling service to run in production environments without concern.

# Big Picture
![hermes_architecture](https://github.com/canonical/hermes/assets/49406051/3e395c68-fcea-4851-9a1b-fd44b84e22ea)

To enhance flexibility, we've modularized the functionalities into distinct elements:
- Collector
  - Deploy this component on the target machine for gathering performance metrics as defined in the configurations. Collected data is stored in /var/log/collector/.
- Parser
  - Transform the collected data into a format easily readable by the frontend.
- Frontend
  - Provide multiple ways to visualize data for analysis.

# Usage
> Using Snap is a convenient installation method for the tool. 
1.  Gather performance metrics on the target machines
```
hermes-time-series-profiling.collector
```
2.  Start a web server to offer RESTful APIs
```
hermes-time-series-profiling.webserver
```
3.  Establish a connection to the web server
- Web
	- Access the web UI at `http://<-webserver_ip->:8080/`.
![hermes_web_ui_frontend](https://github.com/canonical/hermes/assets/49406051/f593e37f-779a-4901-ac30-9c69f45936e0)

- Grafana App
	- Build the Grafana app with `make grafana` and start the Grafana server using `docker-compose up`.
	- Access the Grafana app at http://<grafana_ip>:3000/ and set the web server's IP as the data source.
![hermes_grafana_frontend](https://github.com/canonical/hermes/assets/49406051/1d14729b-cff9-42c1-ab41-8a7e6129abfd)

# Build
Two options:
- Local build
```
sudo make && sudo make install
```
- Snap build
```
snapcraft && sudo snap install hermes_X.X_XXX.snap --dangerous --classic
```
