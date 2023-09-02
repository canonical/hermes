# Introduction
Hermes is a versatile time-series profiling tool for comprehensive system performance analysis. It runs periodic or oneshot jobs defined in the configurations to monitor performance metrics. If system loading stays below defined thresholds, metric collection is skipped to ensure minimal impact on system performance. This allows the profiling service to run in production environments without concern.

# Big Picture
![hermes_architecture](https://github.com/canonical/hermes/assets/49406051/3e395c68-fcea-4851-9a1b-fd44b84e22ea)

To enhance flexibility, we've modularized the functionalities into distinct elements:
- Collector
Deploy this component on the target machine for gathering performance metrics as defined in the configurations. Collected data is stored in /var/log/collector/.
- Parser
Transform the collected data into a format easily readable by the frontend.
- Frontend
Provide two ways to visualize data for analysis - a web UI and a Grafana app.

# Building
You have two options for building the application:
- Local build
```
git clone git@github.com:canonical/hermes.git
cd hermes
# Obtain root access
sudo -i
make && make install
```
- Snap build
```
git clone git@github.com:canonical/hermes.git
cd hermes
snapcraft
# Obtain root access
sudo -i
snap install hermes_1.0_XXX.snap --dangerous --classic
```

# Running
### Configure jobs
Control the collection logic using configurations (/root/hermes/config or /root/snap/hermes/common/hermes/config). You can disable jobs using the {hermes.}job-config command, and you can also modify task parameters within a job by directly overwriting them in the job's YAML.
### Collecting metrics
Collect performance metrics with the {hermes.}collector command. It can also detect configuration changes and respond instantly.
### Parsing metrics
Use the {hermes.}parser command to convert raw data into a frontend-readable format. The parser has two modes:
- Oneshot mode: Suitable for batch processing of stored raw data.
- Daemon mode: Uses pubsub to communicate with the collector, processing new raw data as it becomes available.
### Frontend
Choose between two approaches for performance analysis: web UI or Grafana app. Regardless of your preference, run the {hermes.}webserver command to provide a RESTful API.
- Access the web UI via `http://<-webserver_ip->:8080/`.
![hermes_web_ui_frontend](https://github.com/canonical/hermes/assets/49406051/f593e37f-779a-4901-ac30-9c69f45936e0)

- Set up the Grafana app by executing `docker compose up` command in the hermes/grafana_app folder. Access it at `http://<-grafana_ip->:3000/` for login.
![hermes_grafana_frontend](https://github.com/canonical/hermes/assets/49406051/1d14729b-cff9-42c1-ab41-8a7e6129abfd)
