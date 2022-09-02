package core

import "time"

type JPNICCert struct {
	ID        int
	Name      string
	IsActive  bool
	IsIPv6    bool
	ASN       int
	CA        string
	P12Base64 string
	P12Pass   string
	CAPath    string
	CertPath  string
	P12Path   string
	KeyPath   string
}

//	requestStr += "&deli_no=" + aplyId
//	requestStr += "&rtn_date=" + aplyId
//	requestStr += "&aply_from_addr=" + aplyId
//	requestStr += "&aply_from_addr_confirm=" + aplyId

type Assignment struct {
	IPAddr        string   `json:"ip_addr" res:"ipaddr"`
	NetworkName   string   `json:"network_name" res:"netwrk_nm"`
	InfraUserKind string   `json:"infra_user_kind" res:"infra_usr_kind"`
	Org           string   `json:"org" res:"org_nm_jp"`
	OrgEn         string   `json:"org_en" res:"org_nm"`
	Postcode      string   `json:"postcode" res:"zipcode"`
	Address       string   `json:"address" res:"addr_jp"`
	AddressEn     string   `json:"address_en" res:"addr"`
	AdminHandle   string   `json:"admin_handle" res:"adm_hdl"`
	TechHandle    string   `json:"tech_handle" res:"tech_hdl"`
	NameServer    []string `json:"name_server" res:"nmsrv"`
	NotifyEmail   string   `json:"notify_email" res:"ntfy_mail"`
	Plan          string   `json:"plan" res:"plan_data"`
	DeliNo        string   `json:"deli_no" res:"deli_no"`
	ReturnDate    string   `json:"return_date" res:"rtn_date"`
	ApplyEmail    string   `json:"apply_email" res:"aply_from_addr"`
	Emp           []Emp    `json:"emp"`
}

type Emp struct {
	Kind        string `json:"kind" res:"kind"` // group
	JPNICHandle string `json:"jpnic_handle" res:"jpnic_hdl"`
	Name        string `json:"name" res:"name_jp"`
	NameEn      string `json:"name_en" res:"name"`
	Email       string `json:"email" res:"email"`
	Org         string `json:"org" res:"org_nm_jp"`
	OrgEn       string `json:"org_en" res:"org_nm"`
	Postcode    string `json:"postcode" res:"zipcode"`
	Address     string `json:"address" res:"addr_jp"`
	AddressEn   string `json:"address_en" res:"addr"`
	Division    string `json:"division" res:"division_jp"`
	DivisionEn  string `json:"division_en" res:"division"`
	Title       string `json:"title" res:"title_jp"`
	TitleEn     string `json:"title_en" res:"title"`
	Tel         string `json:"tel" res:"phone"`
	Fax         string `json:"fax" res:"fax"`
	NotifyEmail string `json:"notify_email" res:"ntfy_mail"`
}

type Abuse struct {
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
		IP   string `yaml:"ip"`
		Port uint   `yaml:"port"`
		Name string `yaml:"name"`
		User string `yaml:"user"`
		Pass string `yaml:"pass"`
	} `yaml:"db"`
}
