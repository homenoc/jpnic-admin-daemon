package main

import "time"

type JPNICCert struct {
	ID       int
	Name     string
	IsActive bool
	IsIPv6   bool
	ASN      int
	CAPath   string
	CertPath string
	KeyPath  string
	P12Pass  string
	P12Path  string
}

type ResultV4List struct {
	ID                 int
	GetTime            time.Time
	IPAddress          string
	Size               int
	NetworkName        string
	AssignDate         time.Time
	ReturnDate         time.Time
	Org                string
	OrgEn              string
	ResourceAdminShort string
	RecepNumber        string
	DeliNumber         string
	Types              string
	Division           string
	PostCode           string
	Address            string
	AddressEn          string
	NameServer         string
	DsRecord           string
	NotifyAddress      string
	Asn                string
	AdminJPNIC         string
}

type JPNICHandle struct {
	ID          int
	JPNICHandle string
	GetTime     time.Time
	Name        string
	NameEn      string
	Email       string
	Org         string
	OrgEn       string
	Division    string
	DivisionEn  string
	IsIPv6      string
	Asn         string
}

type Config struct {
	NextTime uint `yaml:"next_time"`
	DB       struct {
		Type string `yaml:"type"`
		Path string `yaml:"path"`
	} `yaml:"db"`
}
