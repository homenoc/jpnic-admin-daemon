package database

type JPNIC struct {
	ID                 uint
	Name               string
	IsActive           bool
	IsIPv6             bool
	Ada                bool
	CollectionInterval uint
	ASN                uint
	P12Base64          string
	P12Pass            string
}

type JPNICHandle struct {
	ID           uint
	GetStartDate string
	GetDate      string
	IsIpv6       bool
	JPNICHandle  string `gorm:"column:jpnic_handle"`
	Name         string
	NameEn       string
	Email        string
	Org          string
	OrgEn        string
	Division     string
	DivisionEn   string
	Tel          string
	Fax          string
	UpdateDate   string
	ASN          uint `gorm:"column:asn_id"`
}

type V4List struct {
	ID                 uint
	GetStartDate       string
	GetDate            string
	IpAddress          string
	Size               uint
	NetworkName        string
	AssignDate         string
	ReturnDate         *string
	Org                string
	OrgEn              string
	ResourceAdminShort string
	RecepNumber        string
	DeliNumber         string
	Type               string
	Division           string
	PostCode           string
	Address            string
	AddressEn          string
	NameServer         string
	DsRecord           string
	NotifyAddress      string
	AdminJpnicId       *uint
	AsnId              uint
}

type V6List struct {
	ID                 uint
	GetStartDate       string
	GetDate            string
	IpAddress          string
	NetworkName        string
	AssignDate         string
	ReturnDate         *string
	Org                string
	OrgEn              string
	ResourceAdminShort string
	RecepNumber        string
	DeliNumber         string
	Division           string
	PostCode           string
	Address            string
	AddressEn          string
	NameServer         string
	DsRecord           string
	NotifyAddress      string
	AdminJpnicId       *uint
	AsnId              uint
}

type V4TechJPNICLists struct {
	ID            uint
	V4ListId      uint `gorm:"column:v4list_id"`
	JpnicHandleId uint `gorm:"column:jpnichandle_id"`
}

type V6TechJPNICLists struct {
	ID            uint
	V6ListId      uint `gorm:"column:v6list_id"`
	JpnicHandleId uint `gorm:"column:jpnichandle_id"`
}
