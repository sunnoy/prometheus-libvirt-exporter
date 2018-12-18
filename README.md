# Prometheus libvirt exporter

## this is fork repo

this repo fork [prometheus-libvirt-exporter](https://github.com/zhangjianweibj/prometheus-libvirt-exporter)

## libvirt api

The libvirt api is [here](https://libvirt.org/html/)

## libvirt go package

This exporter makes use of
[libvirt-go](https://github.com/digitalocean/go-libvirt), the digitalocean Go
bindings for libvirt. 

## export metrics

### base in forks repo
The following metrics/labels are being exported:

Name | Description
---------|-------------
up|scraping libvirt's metrics state
domains_number|get number of domains
domain_state_code|code of the domain state,include state description
maximum_memory_bytes|Maximum allowed memory of the domain, in bytes
memory_usage_bytes|Memory usage of the domain, in bytes
virtual_cpus|Number of virtual CPUs for the domain
cpu_time_seconds_total|Amount of CPU time used by the domain, in seconds
read_bytes_total|Number of bytes read from a block device, in bytes
read_requests_total|Number of read requests from a block device
write_bytes_total|Number of bytes written from a block device, in bytes
write_requests_total|Number of write requests from a block device
receive_bytes_total|Number of bytes received on a network interface, in bytes
receive_packets_total|Number of packets received on a network interface
receive_errors_total|Number of packet receive errors on a network interface
receive_drops_total|Number of packet receive drops on a network interface
transmit_bytes_total|Number of bytes transmitted on a network interface, in bytes
transmit_packets_total|Number of packets transmitted on a network interface
transmit_errors_total|Number of packet transmit errors on a network interface
transmit_drops_total|Number of packet transmit drops on a network interface

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

more can look `output-sample.txt`

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
