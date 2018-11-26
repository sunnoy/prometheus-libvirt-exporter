package libvirt_schema

import (
	"encoding/xml"
)

type Domain struct {
	Devices Devices `xml:"devices"`
	Name string `xml:"name"`
	UUID string `xml:"uuid"`
	Metadata Metadata `xml:"metadata"`
}

type Metadata struct {
	NovaInstance NovaInstance `xml:"instance"`
}

type NovaInstance struct {
	XMLName xml.Name `xml:"instance"`
	Name string `xml:"name"`
}

type Devices struct {
	Disks      []Disk      `xml:"disk"`
	Interfaces []Interface `xml:"interface"`
}

type Disk struct {
	Device string     `xml:"device,attr"`
	Source DiskSource `xml:"source"`
	Target DiskTarget `xml:"target"`
}

type DiskSource struct {
	File string `xml:"file,attr"`
	Protocol string `xml:"protocol,attr"`
	Name string `xml:"name,attr"`
	Host []Host `xml:"host"`


}

type Host struct {
	Name string `xml:"name,attr"`
}

type DiskTarget struct {
	Device string `xml:"dev,attr"`
}

type Interface struct {
	Source InterfaceSource `xml:"source"`
	Target InterfaceTarget `xml:"target"`
	Mac DomainInterfaceMacXml `xml:"mac"`
	Alias DomainInterfaceAliasXml `xml:"alias"`

}

type InterfaceSource struct {
	Bridge string `xml:"bridge,attr"`
}

type InterfaceTarget struct {
	Device string `xml:"dev,attr"`
}


// MAC地址
type DomainInterfaceMacXml struct {
	Address string `xml:"address,attr"`
}


// 网卡AliasName
type DomainInterfaceAliasXml struct{
	Name string `xml:"name,attr"`
}

