# Systemd Unit

_The below instructions mimic Debian's conventions for Prometheus exporters._

The unit file `prometheus-powervault-me5-exporter.service` in this directory should be placed in  `/etc/systemd/system`.

The service runs as a unprivileged user named `prometheus`.

The service references a defaults file located at `/etc/default/prometheus-powervault-me5-exporter`. This file contains the command line arguements and environment variables needed to run the exporter. A sample is found in `powervault_me5_exporter.defaults`.