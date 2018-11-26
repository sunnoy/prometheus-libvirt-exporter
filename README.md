# Prometheus libvirt exporter

## this is fork repo

this repo fork [prometheus-libvirt-exporter](https://github.com/zhangjianweibj/prometheus-libvirt-exporter)

## libvirt api

The libvirt api is [here](https://libvirt.org/html/)

## libvirt go package

This exporter makes use of
[libvirt-go](https://github.com/digitalocean/go-libvirt), the digitalocean Go
bindings for libvirt. Ideally, this exporter should make use of the
`GetAllDomainStats()` API call to extract all relevant metrics.
Unfortunately, we at Kumina still need this exporter to be compatible
with older versions of libvirt that don't support this API call.

## export metrics

### base in forks repo
The following metrics/labels are being exported:

```bash
libvirt_domain_block_stats_read_bytes_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_read_requests_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_write_bytes_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_block_stats_write_requests_total{domain="...",source_file="...",target_device="..."}
libvirt_domain_info_cpu_time_seconds_total{domain="..."}
libvirt_domain_info_maximum_memory_bytes{domain="..."}
libvirt_domain_info_memory_usage_bytes{domain="..."}
libvirt_domain_info_virtual_cpus{domain="..."}
libvirt_domain_interface_stats_receive_bytes_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_drops_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_errors_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_receive_packets_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_bytes_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_drops_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_errors_total{domain="...",source_bridge="...",target_device="..."}
libvirt_domain_interface_stats_transmit_packets_total{domain="...",source_bridge="...",target_device="..."}
libvirt_up
```

### what i do

add block info metrics

```bash
# HELP libvirt_domain_block_info_allocation host physical size in bytes of the image container, in bytes.
libvirt_domain_block_info_allocation{domain="...",source_file="",target_device="..."} 
# HELP libvirt_domain_block_info_capacity how much storage the guest will see, in bytes.
libvirt_domain_block_info_capacity{domain="...",source_file="",target_device="..."} 
# HELP libvirt_domain_block_info_physical host storage in bytes occupied by the image, in bytes.
libvirt_domain_block_info_physical{domain="...",source_file="",target_device="..."} 

#some momory metrics
virsh dommemstat test
actual 1048576
swap_in 0
swap_out 0
major_fault 484
minor_fault 48658488
unused 797544
available 1016008
usable 767732
last_update 1543212082
rss 386304
```

## build

about the project look [here](https://www.li-rui.top/2018/11/21/monitor/%E5%BC%80%E5%8F%91libvirt_exporter%E7%9A%84go%E7%89%88%E6%9C%AC/)

## how to use

This repository provides code for a Prometheus metrics exporter
for [libvirt](https://libvirt.org/). This exporter connects to any
libvirt daemon and exports per-domain metrics related to CPU, memory,
disk and network usage. By default, this exporter listens on TCP port
9000 and use libvirtd api by unix remote api (just listen libvirtd sock)

after build ，you can to see hlep

```bash
./prometheus-libvirt-exporter  -h
Usage of ./prometheus-libvirt-exporter:
  -libvirt.uri string
        Libvirt URI from which to extract metrics. (default "/var/run/libvirt/libvirt-sock")
  -web.listen-address string
        Address to listen on for web interface and telemetry. (default ":9000")
  -web.telemetry-path string
        Path under which to expose metrics. (default "/metrics")

```
