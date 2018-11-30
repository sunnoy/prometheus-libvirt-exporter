package main

import (
	"encoding/xml"
	"flag"
	"github.com/digitalocean/go-libvirt"
	"github.com/prometheus/client_golang/prometheus"
	"log"
	"net"
	"net/http"
	"prometheus-libvirt-exporter/libvirt_schema"
	"time"
)

var (
	libvirtUpDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "", "up"),
		"Whether scraping libvirt's metrics was successful.",
		[]string{"host"},
		nil)

	libvirtDomainNumbers = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "", "domains_number"),
		"Number of the domain",
		[]string{"host"},
		nil)

	libvirtDomainState = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "", "domain_state_code"),
		"Code of the domain state",
		[]string{"domain", "instanceName", "instanceId", "stateDesc", "host"},
		nil)

	libvirtDomainInfoMaxMemDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "maximum_memory_bytes"),
		"Maximum allowed memory of the domain, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainInfoMemoryDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "memory_usage_bytes"),
		"Memory usage of the domain, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainInfoNrVirtCpuDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "virtual_cpus"),
		"Number of virtual CPUs for the domain.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainInfoCpuTimeDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_info", "cpu_time_seconds_total"),
		"Amount of CPU time used by the domain, in seconds.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)

	libvirtDomainBlockRdBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_bytes_total"),
		"Number of bytes read from a block device, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "source_file", "target_device", "host"},
		nil)
	libvirtDomainBlockRdReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "read_requests_total"),
		"Number of read requests from a block device.",
		[]string{"domain", "instanceName", "instanceId", "source_file", "target_device", "host"},
		nil)
	libvirtDomainBlockWrBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_bytes_total"),
		"Number of bytes written from a block device, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "source_file", "target_device", "host"},
		nil)
	libvirtDomainBlockWrReqDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_stats", "write_requests_total"),
		"Number of write requests from a block device.",
		[]string{"domain", "instanceName", "instanceId", "source_file", "target_device", "host"},
		nil)

	//DomainInterface
	libvirtDomainInterfaceRxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_bytes_total"),
		"Number of bytes received on a network interface, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)
	libvirtDomainInterfaceRxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_packets_total"),
		"Number of packets received on a network interface.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)
	libvirtDomainInterfaceRxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_errors_total"),
		"Number of packet receive errors on a network interface.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)
	libvirtDomainInterfaceRxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "receive_drops_total"),
		"Number of packet receive drops on a network interface.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)
	libvirtDomainInterfaceTxBytesDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_bytes_total"),
		"Number of bytes transmitted on a network interface, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)
	libvirtDomainInterfaceTxPacketsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_packets_total"),
		"Number of packets transmitted on a network interface.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)
	libvirtDomainInterfaceTxErrsDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_errors_total"),
		"Number of packet transmit errors on a network interface.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)
	libvirtDomainInterfaceTxDropDesc = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_interface_stats", "transmit_drops_total"),
		"Number of packet transmit drops on a network interface.",
		[]string{"domain", "instanceName", "instanceId", "source_bridge", "target_device", "mac_address", "host"},
		nil)

	domainState = map[libvirt_schema.DomainState]string{
		libvirt_schema.DOMAIN_NOSTATE:     "no state",
		libvirt_schema.DOMAIN_RUNNING:     "the domain is running",
		libvirt_schema.DOMAIN_BLOCKED:     "the domain is blocked on resource",
		libvirt_schema.DOMAIN_PAUSED:      "the domain is paused by user",
		libvirt_schema.DOMAIN_SHUTDOWN:    "the domain is being shut down",
		libvirt_schema.DOMAIN_SHUTOFF:     "the domain is shut off",
		libvirt_schema.DOMAIN_CRASHED:     "the domain is crashed",
		libvirt_schema.DOMAIN_PMSUSPENDED: "the domain is suspended by guest power management",
		libvirt_schema.DOMAIN_LAST:        "this enum value will increase over time as new events are added to the libvirt API",
	}

	// BlockCapacity
	libvirtDomainBlockCapacity = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_info", "capacity"),
		"how much storage the guest will see, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "source_file", "source_rbd", "source_ceph_mon_host", "target_device", "host"},
		nil)

	//BlockAllocation
	libvirtDomainBlockAllocation = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_info", "allocation"),
		"host storage in bytes occupied by the image, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "source_file", "source_rbd", "source_ceph_mon_host", "target_device", "host"},
		nil)

	//BlockPhysical
	libvirtDomainBlockPhysical = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_block_info", "physical"),
		"host physical size in bytes of the image container, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "source_file", "source_rbd", "source_ceph_mon_host", "target_device", "host"},
		nil)

	// memory static
	libvirtDomainMemorySwapin = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "swapin"),
		"guest memory swapin, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemorySwapout = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "swapout"),
		"guest memory swapout, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryMajorfault = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "majorfault"),
		"guest memory majorfault, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryMinorfault = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "minorfault"),
		"guest memory minorfault, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryUnused = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "unused"),
		"guest memory unused, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryAvailable = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "available"),
		"guest memory available, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryActualballon = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "actualballon"),
		"guest memory actualballon, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryRss = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "rss"),
		"guest memory rss, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryUsable = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "usable"),
		"guest memory usable, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryLastupdate = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "lastupdate"),
		"guest memory lastupdate, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
	libvirtDomainMemoryNr = prometheus.NewDesc(
		prometheus.BuildFQName("libvirt", "domain_memory_info", "nr"),
		"guest memory nr, in bytes.",
		[]string{"domain", "instanceName", "instanceId", "host"},
		nil)
)

// CollectDomain extracts Prometheus metrics from a libvirt domain.
func CollectDomain(ch chan<- prometheus.Metric, l *libvirt.Libvirt, domain *libvirt.Domain) error {
	xmlDesc, err := l.DomainGetXMLDesc(*domain, 0)
	if err != nil {
		log.Fatalf("failed to DomainGetXMLDesc: %v", err)
		return err
	}
	var libvirtSchema libvirt_schema.Domain
	//Unmarshal make the data of xmlDesc into the struct libvirtSchema
	err = xml.Unmarshal([]byte(xmlDesc), &libvirtSchema)
	if err != nil {
		log.Fatalf("failed to Unmarshal domains: %v", err)
		return err
	}

	domainName := domain.Name
	instanceName := libvirtSchema.Metadata.NovaInstance.Name
	instanceId := libvirtSchema.UUID
	host, err := l.ConnectGetHostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
		return err
	}

	rState, rmaxmem, rmemory, rvirCpu, rcputime, err := l.DomainGetInfo(*domain)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainState,
		prometheus.GaugeValue,
		float64(rState),
		domainName, instanceName, instanceId, domainState[libvirt_schema.DomainState(rState)], host)

	if err != nil {
		log.Fatalf("failed to get domainInfo: %v", err)
		return err
	}
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoMaxMemDesc,
		prometheus.GaugeValue,
		float64(rmaxmem)*1024,
		domainName, instanceName, instanceId, host)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoMemoryDesc,
		prometheus.GaugeValue,
		float64(rmemory)*1024,
		domainName, instanceName, string(instanceId[:]), host)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoNrVirtCpuDesc,
		prometheus.GaugeValue,
		float64(rvirCpu),
		domainName, instanceName, string(instanceId[:]), host)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainInfoCpuTimeDesc,
		prometheus.CounterValue,
		float64(rcputime)/1e9,
		domainName, instanceName, string(instanceId[:]), host)

	isActive, _ := l.DomainIsActive(*domain)
	if isActive == 1 {

		//report memory statistics
		rDomainMemroyStats, rgetDomainMemoryStatsErr := l.DomainMemoryStats(*domain, 11, 0)
		if (rgetDomainMemoryStatsErr != nil) {
			log.Fatalf("error getting instance memory state  of virsh dommemstat for '%s': %v", domainName, rgetDomainMemoryStatsErr)
		}

		//get domains memory
		for _, domainMemoryState := range rDomainMemroyStats {

			switch domainMemoryState.Tag {
			case 0:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemorySwapin,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val)*1024,
					domainName, instanceName, string(instanceId[:]), host)

			case 1:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemorySwapout,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val)*1024,
					domainName, instanceName, string(instanceId[:]), host)

			case 2:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryMajorfault,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val),
					domainName, instanceName, string(instanceId[:]), host)

			case 3:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryMinorfault,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val),
					domainName, instanceName, string(instanceId[:]), host)

			case 4:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryUnused,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val)*1024,
					domainName, instanceName, string(instanceId[:]), host)

			case 5:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryAvailable,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val)*1024,
					domainName, instanceName, string(instanceId[:]), host)

			case 6:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryActualballon,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val)*1024,
					domainName, instanceName, string(instanceId[:]), host)

			case 7:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryRss,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val)*1024,
					domainName, instanceName, string(instanceId[:]), host)
				//return "rss", value * 1024	//qemu process在宿主机上所占用的内存，可以通过 grep VmRSS /proc/$(pidof qemu-system-x86_64)/status 得到
			case 8:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryUsable,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val)*1024,
					domainName, instanceName, string(instanceId[:]), host)
				//return "usable", value * 1024	// How much the balloon can be inflated without pushing the guest system to swap, corresponds to 'Available' in /proc/meminfo
			case 9:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryLastupdate,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val),
					domainName, instanceName, string(instanceId[:]), host)
				//return "lastupdate", value	//Timestamp of the last statistic
			case 10:
				ch <- prometheus.MustNewConstMetric(
					libvirtDomainMemoryNr,
					prometheus.GaugeValue,
					float64(domainMemoryState.Val),
					domainName, instanceName, string(instanceId[:]), host)
				//return "nr", value

			}
		}

	}

	// Report block device statistics.
	for _, disk := range libvirtSchema.Devices.Disks {
		if disk.Device == "cdrom" || disk.Device == "fd" {
			continue
		}

		//fmt.Print(disk.Source.Host)
		//fmt.Print(disk.Source.Name)
		//os.Exit(3)

		isActive, err := l.DomainIsActive(*domain)

		var rRdReq, rRdBytes, rWrReq, rWrBytes int64
		var rAllocation, rCapacity, rPhysical uint64
		var Sname, Shost string

		if isActive == 1 {
			rRdReq, rRdBytes, rWrReq, rWrBytes, _, err = l.DomainBlockStats(*domain, disk.Target.Device)
			rAllocation, rCapacity, rPhysical, err = l.DomainGetBlockInfo(*domain, disk.Target.Device, 0)
		}

		if err != nil {
			log.Fatalf("failed to get DomainBlockStats: %v", err)
			return err
		}


		if disk.Source.Name != "" {
			Sname = disk.Source.Name
		}

		if disk.Source.Host != nil {
			Shost = disk.Source.Host[0].Name
		}
		////block info
		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockCapacity,
			prometheus.GaugeValue,
			float64(rCapacity),
			domainName, instanceName, string(instanceId[:]),
			disk.Source.File,
			//rbd
			Sname,
			//mon host
			Shost,
			disk.Target.Device,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockAllocation,
			prometheus.GaugeValue,
			float64(rAllocation),
			domainName, instanceName, string(instanceId[:]),
			disk.Source.File,
			//rbd
			Sname,
			//mon host
			Shost,
			disk.Target.Device,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockPhysical,
			prometheus.GaugeValue,
			float64(rPhysical),
			domainName, instanceName, string(instanceId[:]),
			disk.Source.File,
			//rbd
			Sname,
			//mon host
			Shost,
			disk.Target.Device,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockRdBytesDesc,
			prometheus.CounterValue,
			float64(rRdBytes),
			domainName, instanceName, string(instanceId[:]),
			disk.Source.File,
			disk.Target.Device,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockRdReqDesc,
			prometheus.CounterValue,
			float64(rRdReq),
			domainName, instanceName, string(instanceId[:]),
			disk.Source.File,
			disk.Target.Device,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockWrBytesDesc,
			prometheus.CounterValue,
			float64(rWrBytes),
			domainName, instanceName, string(instanceId[:]),
			disk.Source.File,
			disk.Target.Device,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainBlockWrReqDesc,
			prometheus.CounterValue,
			float64(rWrReq),
			domainName, instanceName, string(instanceId[:]),
			disk.Source.File,
			disk.Target.Device,
			host)

	}

	// Report network interface statistics.
	for _, iface := range libvirtSchema.Devices.Interfaces {
		if iface.Target.Device == "" {
			continue
		}

		//fmt.Print(iface.Mac.Address)
		//fmt.Print(iface.Alias.Name)
		//os.Exit(3)

		isActive, err := l.DomainIsActive(*domain)
		var rRxBytes, rRxPackets, rRxErrs, rRxDrop, rTxBytes, rTxPackets, rTxErrs, rTxDrop int64

		if isActive == 1 {
			rRxBytes, rRxPackets, rRxErrs, rRxDrop, rTxBytes, rTxPackets, rTxErrs, rTxDrop, err = l.DomainInterfaceStats(*domain, iface.Target.Device)

		}

		if err != nil {
			log.Fatalf("failed to get DomainInterfaceStats: %v", err)
			return err
		}

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxBytesDesc,
			prometheus.CounterValue,
			float64(rRxBytes),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxPacketsDesc,
			prometheus.CounterValue,
			float64(rRxPackets),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxErrsDesc,
			prometheus.CounterValue,
			float64(rRxErrs),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceRxDropDesc,
			prometheus.CounterValue,
			float64(rRxDrop),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxBytesDesc,
			prometheus.CounterValue,
			float64(rTxBytes),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxPacketsDesc,
			prometheus.CounterValue,
			float64(rTxPackets),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxErrsDesc,
			prometheus.CounterValue,
			float64(rTxErrs),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

		ch <- prometheus.MustNewConstMetric(
			libvirtDomainInterfaceTxDropDesc,
			prometheus.CounterValue,
			float64(rTxDrop),
			domainName, instanceName, string(instanceId[:]),
			iface.Source.Bridge,
			iface.Target.Device,
			iface.Mac.Address,
			host)

	}

	return nil
}



// CollectFromLibvirt obtains Prometheus metrics from all domains in a
// libvirt setup.
func CollectFromLibvirt(ch chan<- prometheus.Metric, uri string) error {
	conn, err := net.DialTimeout("unix", uri, 5*time.Second)
	if err != nil {
		log.Fatalf("failed to dial libvirt: %v", err)
		return err
	}
	defer conn.Close()

	l := libvirt.New(conn)
	if err = l.Connect(); err != nil {
		log.Fatalf("failed to connect: %v", err)
		return err
	}

	host, err := l.ConnectGetHostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
		return err
	}

	ch <- prometheus.MustNewConstMetric(
		libvirtUpDesc,
		prometheus.GaugeValue,
		1.0,
		host)

	domains, err := l.Domains()
	if err != nil {
		log.Fatalf("failed to load domain: %v", err)
		return err
	}

	//domains number
	domainNumber := len(domains)
	ch <- prometheus.MustNewConstMetric(
		libvirtDomainNumbers,
		prometheus.GaugeValue,
		float64(domainNumber),
		host)

	for _, domain := range domains {
		err = CollectDomain(ch, l, &domain)
		//sb code delete
		//l.DomainShutdown(domain)
		//domain.Free()
		if err != nil {
			log.Fatalf("failed to Collect domain: %v", err)
			return err
		}
	}
	return nil
}

// LibvirtExporter implements a Prometheus exporter for libvirt state.
type LibvirtExporter struct {
	uri string
}

// NewLibvirtExporter creates a new Prometheus exporter for libvirt.
func NewLibvirtExporter(uri string) (*LibvirtExporter, error) {
	return &LibvirtExporter{
		uri: uri,
	}, nil
}

// Describe returns metadata for all Prometheus metrics that may be exported.
func (e *LibvirtExporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- libvirtUpDesc
	ch <- libvirtDomainNumbers

	//domain info
	ch <- libvirtDomainState
	ch <- libvirtDomainInfoMaxMemDesc
	ch <- libvirtDomainInfoMemoryDesc
	ch <- libvirtDomainInfoNrVirtCpuDesc
	ch <- libvirtDomainInfoCpuTimeDesc

	//domain block
	ch <- libvirtDomainBlockRdBytesDesc
	ch <- libvirtDomainBlockRdReqDesc
	ch <- libvirtDomainBlockWrBytesDesc
	ch <- libvirtDomainBlockWrReqDesc

	//domain block info
	ch <- libvirtDomainBlockCapacity
	ch <- libvirtDomainBlockAllocation
	ch <- libvirtDomainBlockPhysical

	//domain memory info
	ch <- libvirtDomainMemorySwapin
	ch <- libvirtDomainMemorySwapout
	ch <- libvirtDomainMemoryMinorfault
	ch <- libvirtDomainMemoryMajorfault
	ch <- libvirtDomainMemoryUnused
	ch <- libvirtDomainMemoryAvailable
	ch <- libvirtDomainMemoryActualballon
	ch <- libvirtDomainMemoryRss
	ch <- libvirtDomainMemoryUsable
	ch <- libvirtDomainMemoryLastupdate
	ch <- libvirtDomainMemoryNr

	//domain interface
	ch <- libvirtDomainInterfaceRxBytesDesc
	ch <- libvirtDomainInterfaceRxPacketsDesc
	ch <- libvirtDomainInterfaceRxErrsDesc
	ch <- libvirtDomainInterfaceRxDropDesc
	ch <- libvirtDomainInterfaceTxBytesDesc
	ch <- libvirtDomainInterfaceTxPacketsDesc
	ch <- libvirtDomainInterfaceTxErrsDesc
	ch <- libvirtDomainInterfaceTxDropDesc

}

// Collect scrapes Prometheus metrics from libvirt.
func (e *LibvirtExporter) Collect(ch chan<- prometheus.Metric) {
	CollectFromLibvirt(ch, e.uri)
}

func main() {
	var (
		listenAddress = flag.String("web.listen-address", ":9000", "Address to listen on for web interface and telemetry.")
		metricsPath   = flag.String("web.telemetry-path", "/metrics", "Path under which to expose metrics.")
		libvirtURI    = flag.String("libvirt.uri", "/var/run/libvirt/libvirt-sock", "Libvirt URI from which to extract metrics.")
	)
	flag.Parse()

	exporter, err := NewLibvirtExporter(*libvirtURI)
	if err != nil {
		panic(err)
	}
	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`
			<html>
			<head><title>Libvirt Exporter</title></head>
			<body>
			<h1>Libvirt Exporter</h1>
			<p><a href='` + *metricsPath + `'>Metrics</a></p>
			</body>
			</html>`))
	})
	log.Fatal(http.ListenAndServe(*listenAddress, nil))

}
