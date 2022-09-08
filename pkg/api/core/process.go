package core

import (
	"fmt"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/database"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/jpnic"
	"github.com/homenoc/jpnic-admin-daemon/pkg/api/core/tool/etc"
	"log"
	"net"
	"strconv"
	"strings"
)

type Process interface {
	getBaseIPList() error
	getIPv6() bool
	checkJPNICDataExistsCounter() (int64, int64, error)
	irregularProcess(uint) (bool, error)
	emptyCheck() bool
	getBaseData() error
	irregularCheck() error
	getSameRecepNumID() ([]uint, error)
	getDetail(strHandles []string, handles map[string]uint, ids []uint) error
}

type ipv4 struct {
	base       base
	baseV4List database.V4List
}

type ipv6 struct {
	base       base
	baseV6List database.V6List
}

func (v4 *ipv4) getBaseIPList() error {
	list, err := v4.base.db.GetRangeV4List(v4.base.todayStartTime, v4.base.todayEndTime, v4.base.cert.Base.ID)
	if err != nil {
		return err
	}
	v4.baseV4List = list
	return nil
}

func (v6 *ipv6) getBaseIPList() error {
	list, err := v6.base.db.GetRangeV6List(v6.base.todayStartTime, v6.base.todayEndTime, v6.base.cert.Base.ID)
	if err != nil {
		return err
	}
	v6.baseV6List = list
	return nil
}

func (v4 *ipv4) checkJPNICDataExistsCounter() (int64, int64, error) {
	var noGet, got int64
	noGet, err := v4.base.db.GetV4JPNICDataNotReceivedCount(v4.base.todayStartTime, v4.base.todayEndTime, v4.base.cert.Base.ID)
	if err != nil {
		return noGet, got, err
	}

	got, err = v4.base.db.GetV4JPNICDataReceivedCount(v4.base.todayStartTime, v4.base.todayEndTime, v4.base.cert.Base.ID)
	if err != nil {
		return noGet, got, err
	}

	return noGet, got, nil
}

func (v6 *ipv6) checkJPNICDataExistsCounter() (int64, int64, error) {
	var noGet, got int64
	noGet, err := v6.base.db.GetV6JPNICDataNotReceivedCount(v6.base.todayStartTime, v6.base.todayEndTime, v6.base.cert.Base.ID)
	if err != nil {
		return noGet, got, err
	}

	got, err = v6.base.db.GetV6JPNICDataReceivedCount(v6.base.todayStartTime, v6.base.todayEndTime, v6.base.cert.Base.ID)
	if err != nil {
		return noGet, got, err
	}

	return noGet, got, nil
}

func checkJPNICDataExists(p Process) (bool, error) {
	noGet, got, err := p.checkJPNICDataExistsCounter()
	if err != nil {
		return false, err
	}
	// noGet(未取得), got(取得済)
	// noGet(0) got(0)時は初回取得
	// got(0)時は、すべてのデータ取得済み
	//log.Println(noGet, got)
	if noGet == 0 && got > 0 {
		return true, nil
	}
	return false, nil
}

func (v4 ipv4) getIPv6() bool {
	return false
}

func (v6 ipv6) getIPv6() bool {
	return true
}

func (v4 ipv4) irregularProcess(asn uint) (bool, error) {
	if v4.baseV4List.ID > 0 && v4.baseV4List.RecepNumber == "" {
		err := v4.base.db.UpdateV4ListIrregular(v4.baseV4List.ID, etc.GetTimeDate(), asn)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func (v6 ipv6) irregularProcess(asn uint) (bool, error) {
	if v6.baseV6List.ID > 0 && v6.baseV6List.RecepNumber == "" {
		err := v6.base.db.UpdateV6ListIrregular(v6.baseV6List.ID, etc.GetTimeDate(), asn)
		if err != nil {
			return true, err
		}
		return true, nil
	}
	return false, nil
}

func (v4 ipv4) emptyCheck() bool {
	return v4.baseV4List == database.V4List{}
}

func (v6 ipv6) emptyCheck() bool {
	return v6.baseV6List == database.V6List{}
}

func (v4 ipv4) getBaseData() error {
	// 1000件以上の場合も取得
	var infos []jpnic.InfoIPv4
	isOverList := true
	addressRange := ""

	for isOverList {
		filter := jpnic.SearchIPv4{
			IsDetail: false,
			Option1:  nil,
		}
		if addressRange != "" {
			filter.IPAddress = addressRange
		}
		data, err := v4.base.jpnicConfig.SearchIPv4(filter)
		if err != nil {

			return fmt.Errorf("[getBaseData,SearchIPv4] Error:  %s", err)
		}

		isOverList = data.IsOverList

		if isOverList {
			lastIPAddress, _, err := net.ParseCIDR(data.InfoIPv4[len(data.InfoIPv4)-1].IPAddress)
			if err != nil {
				log.Println("Error", "parseCIDR", err)
				return err
			}

			addressRange = lastIPAddress.String() + "-255.255.255.255"
		}

		infos = append(infos, data.InfoIPv4...)
	}

	for _, tmp := range infos {
		sizeNum, _ := strconv.Atoi(tmp.Size)
		var returnDate *string = nil
		if tmp.ReturnDate != "" {
			returnDate = &tmp.ReturnDate
		}
		adminJPNICHandleNum, _ := strconv.Atoi(tmp.InfoDetail.AdminJPNICHandle)
		var adminJpnicId *uint = nil
		if adminJPNICHandleNum != 0 {
			adminJpnicId = &[]uint{uint(adminJPNICHandleNum)}[0]
		}
		_, err := v4.base.db.CreateResultV4List(database.V4List{
			GetStartDate:       etc.GetTimeDate(),
			GetDate:            etc.GetTimeDate(),
			IsDisabled:         false,
			IsGet:              true,
			IpAddress:          tmp.IPAddress,
			Size:               uint(sizeNum),
			NetworkName:        tmp.NetworkName,
			AssignDate:         tmp.AssignDate,
			ReturnDate:         returnDate,
			Org:                tmp.InfoDetail.Org,
			OrgEn:              tmp.InfoDetail.OrgEn,
			ResourceAdminShort: tmp.InfoDetail.Ryakusho,
			RecepNumber:        tmp.RecepNo,
			DeliNumber:         tmp.DeliNo,
			Type:               tmp.Type,
			Division:           tmp.Division,
			PostCode:           tmp.InfoDetail.PostCode,
			Address:            tmp.InfoDetail.Address,
			AddressEn:          tmp.InfoDetail.AddressEn,
			NameServer:         strings.Join(tmp.InfoDetail.NameServer, ","),
			DsRecord:           tmp.InfoDetail.DSRecord,
			NotifyAddress:      tmp.InfoDetail.NotifyAddress,
			AdminJpnicId:       adminJpnicId,
			AsnId:              v4.base.cert.Base.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (v6 ipv6) getBaseData() error {
	// 1000件以上の場合も取得
	var infos []jpnic.InfoIPv6
	isOverList := true
	addressRange := ""

	for isOverList {
		filter := jpnic.SearchIPv6{
			IsDetail: false,
			Option1:  nil,
		}
		if addressRange != "" {
			filter.IPAddress = addressRange
		}
		data, err := v6.base.jpnicConfig.SearchIPv6(filter)
		if err != nil {
			return nil
		}

		isOverList = data.IsOverList

		// Todo: 後ほど実装(1000件を超える場合)
		//if isOverList {
		//	lastIPAddress, _, err := net.ParseCIDR(data.InfoIPv6[len(data.InfoIPv6)-1].IPAddress)
		//	if err != nil {
		//		log.Println(err)
		//	}
		//
		//	addressRange = lastIPAddress.String() + "-255.255.255.255"
		//}

		infos = append(infos, data.InfoIPv6...)
	}

	for _, tmp := range infos {
		var returnDate *string = nil
		if tmp.ReturnDate != "" {
			returnDate = &tmp.ReturnDate
		}
		adminJPNICHandleNum, _ := strconv.Atoi(tmp.InfoDetail.AdminJPNICHandle)
		var adminJpnicId *uint = nil
		if adminJPNICHandleNum != 0 {
			adminJpnicId = &[]uint{uint(adminJPNICHandleNum)}[0]
		}
		_, err := v6.base.db.CreateResultV6List(database.V6List{
			GetStartDate:       etc.GetTimeDate(),
			GetDate:            etc.GetTimeDate(),
			IsDisabled:         false,
			IsGet:              true,
			IpAddress:          tmp.IPAddress,
			NetworkName:        tmp.NetworkName,
			AssignDate:         tmp.AssignDate,
			ReturnDate:         returnDate,
			Org:                tmp.InfoDetail.Org,
			OrgEn:              tmp.InfoDetail.OrgEn,
			ResourceAdminShort: tmp.InfoDetail.Ryakusho,
			RecepNumber:        tmp.RecepNo,
			DeliNumber:         tmp.DeliNo,
			PostCode:           tmp.InfoDetail.PostCode,
			Address:            tmp.InfoDetail.Address,
			AddressEn:          tmp.InfoDetail.AddressEn,
			NameServer:         strings.Join(tmp.InfoDetail.NameServer, ","),
			DsRecord:           tmp.InfoDetail.DSRecord,
			NotifyAddress:      tmp.InfoDetail.NotifyAddress,
			AdminJpnicId:       adminJpnicId,
			AsnId:              v6.base.cert.Base.ID,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func (v4 ipv4) irregularCheck() error {
	if v4.baseV4List.RecepNumber == "" {
		return fmt.Errorf("%s", "data is not found...")
	}
	return nil
}

func (v6 ipv6) irregularCheck() error {
	if v6.baseV6List.RecepNumber == "" {
		return fmt.Errorf("%s", "data is not found...")
	}
	return nil
}

func (v4 ipv4) getSameRecepNumID() ([]uint, error) {
	// 同じ受付番号がないか確認
	var idList []uint
	listsByRecpNum, err := v4.base.db.GetRangeV4ListByRecepNumber(v4.base.todayStartTime, v4.base.todayEndTime, v4.baseV4List.RecepNumber, v4.base.cert.Base.ID)
	if err != nil {
		return idList, err
	}
	for _, listByRecpNum := range listsByRecpNum {
		idList = append(idList, listByRecpNum.ID)
	}
	return idList, nil
}

func (v6 ipv6) getSameRecepNumID() ([]uint, error) {
	// 同じ受付番号がないか確認
	var idList []uint
	listsByRecpNum, err := v6.base.db.GetRangeV6ListByRecepNumber(v6.base.todayStartTime, v6.base.todayEndTime, v6.baseV6List.RecepNumber, v6.base.cert.Base.ID)
	if err != nil {
		return idList, err
	}
	for _, listByRecpNum := range listsByRecpNum {
		idList = append(idList, listByRecpNum.ID)
	}
	return idList, nil
}

func (v4 ipv4) getDetail(strHandles []string, handles map[string]uint, ids []uint) error {
	data, err := v4.base.jpnicConfig.SearchIPv4(jpnic.SearchIPv4{
		IsDetail:  true,
		Option1:   strHandles,
		IPAddress: v4.baseV4List.IpAddress,
		RecepNo:   v4.baseV4List.RecepNumber,
	})
	if err != nil {
		log.Println("Error", "searchIPv4(detail)", err)
		return err
	}

	// jpnic_handle DBに追加処理
	if data != nil && len(data.JPNICHandleDetail) != 0 {
		//log.Println("AdminJPNICHandle", data.InfoIPv4[0].InfoDetail.AdminJPNICHandle)
		//log.Println("TechJPNICHandle", data.InfoIPv4[0].InfoDetail.TechJPNICHandles)
		//log.Println("data.JPNICHandleDetail", data.JPNICHandleDetail)
		for _, jpnicHandle := range data.JPNICHandleDetail {
			//fmt.Printf("%#v\n", jpnicHandle)
			resultJPNICHandle, err := v4.base.db.CreateJPNICHandle(database.JPNICHandle{
				GetStartDate: etc.GetTimeDate(),
				GetDate:      etc.GetTimeDate(),
				IsDisabled:   false,
				IsGet:        false,
				IsIpv6:       false,
				JPNICHandle:  jpnicHandle.JPNICHandle,
				Name:         jpnicHandle.Name,
				NameEn:       jpnicHandle.NameEn,
				Email:        jpnicHandle.Email,
				Org:          jpnicHandle.Org,
				OrgEn:        jpnicHandle.OrgEn,
				Division:     jpnicHandle.Division,
				DivisionEn:   jpnicHandle.DivisionEn,
				Tel:          jpnicHandle.Tel,
				Fax:          jpnicHandle.Fax,
				UpdateDate:   jpnicHandle.UpdateDate,
				ASN:          v4.base.cert.Base.ID,
			})
			if err != nil {
				log.Println("jpnic handle data create:", err)
				return err
			}
			handles[jpnicHandle.JPNICHandle] = resultJPNICHandle.ID
		}
	}

	// Base情報を取得後に返却した場合(JPNIC Handleが消えた場合)
	if data == nil {
		err = v4.base.db.UpdateV4List(ids, database.V4List{
			GetDate:    etc.GetTimeDate(),
			IsDisabled: true,
			IsGet:      false,
			AsnId:      v4.base.cert.Base.ID,
		})
		if err != nil {
			return err
		}
		return nil
	}

	// result_v4list DBにUpdate処理
	err = v4.base.db.UpdateV4List(ids, database.V4List{
		GetDate:       etc.GetTimeDate(),
		IsDisabled:    false,
		IsGet:         false,
		Org:           data.InfoIPv4[0].InfoDetail.Org,
		OrgEn:         data.InfoIPv4[0].InfoDetail.OrgEn,
		PostCode:      data.InfoIPv4[0].InfoDetail.PostCode,
		Address:       data.InfoIPv4[0].InfoDetail.Address,
		AddressEn:     data.InfoIPv4[0].InfoDetail.AddressEn,
		NameServer:    strings.Join(data.InfoIPv4[0].InfoDetail.NameServer, ","),
		DsRecord:      data.InfoIPv4[0].InfoDetail.DSRecord,
		NotifyAddress: data.InfoIPv4[0].InfoDetail.NotifyAddress,
		AdminJpnicId:  &[]uint{handles[data.InfoIPv4[0].InfoDetail.AdminJPNICHandle]}[0],
		AsnId:         v4.base.cert.Base.ID,
	})
	if err != nil {
		log.Println("Error", "update result_jpnichandle", err)
		return err
	}
	// JPNIC技術連絡先をDBに登録
	v4TechJPNICLists := []database.V4TechJPNICLists{}
	for _, id := range ids {
		for _, techJPNICHandle := range data.InfoIPv4[0].InfoDetail.TechJPNICHandles {
			v4TechJPNICLists = append(v4TechJPNICLists, database.V4TechJPNICLists{
				V4ListId:      id,
				JpnicHandleId: handles[techJPNICHandle.TechJPNICHandle],
			})
		}
	}
	_, err = v4.base.db.CreateResultV4ListTechJPNIC(v4TechJPNICLists)
	if err != nil {
		log.Println("Error", "prepare(INSERT) result_v4list_tech_jpnic", err)
		return err
	}

	return nil
}

func (v6 ipv6) getDetail(strHandles []string, handles map[string]uint, ids []uint) error {
	data, err := v6.base.jpnicConfig.SearchIPv6(jpnic.SearchIPv6{
		IsDetail:  true,
		Option1:   strHandles,
		IPAddress: v6.baseV6List.IpAddress,
		RecepNo:   v6.baseV6List.RecepNumber,
	})
	if err != nil {
		return err
	}

	// jpnic_handle DBに追加処理
	if data != nil && len(data.JPNICHandleDetail) != 0 {
		for _, jpnicHandle := range data.JPNICHandleDetail {

			resultJPNICHandle, err := v6.base.db.CreateJPNICHandle(database.JPNICHandle{
				GetStartDate: etc.GetTimeDate(),
				GetDate:      etc.GetTimeDate(),
				IsDisabled:   false,
				IsGet:        false,
				IsIpv6:       false,
				JPNICHandle:  jpnicHandle.JPNICHandle,
				Name:         jpnicHandle.Name,
				NameEn:       jpnicHandle.NameEn,
				Email:        jpnicHandle.Email,
				Org:          jpnicHandle.Org,
				OrgEn:        jpnicHandle.OrgEn,
				Division:     jpnicHandle.Division,
				DivisionEn:   jpnicHandle.DivisionEn,
				Tel:          jpnicHandle.Tel,
				Fax:          jpnicHandle.Fax,
				UpdateDate:   jpnicHandle.UpdateDate,
				ASN:          v6.base.cert.Base.ID,
			})
			if err != nil {
				log.Println("jpnic handle data create:", err)
			}
			handles[jpnicHandle.JPNICHandle] = resultJPNICHandle.ID
		}
	}

	// Base情報を取得後に返却した場合(JPNIC Handleが消えた場合)
	if data == nil {
		err = v6.base.db.UpdateV6List(ids, database.V6List{
			GetDate:    etc.GetTimeDate(),
			IsDisabled: true,
			IsGet:      false,
			AsnId:      v6.base.cert.Base.ID,
		})
		if err != nil {
			return err
		}
		return nil
	}

	// result_v6list DBにUpdate処理
	err = v6.base.db.UpdateV6List(ids, database.V6List{
		GetDate:       etc.GetTimeDate(),
		IsDisabled:    false,
		IsGet:         false,
		Org:           data.InfoIPv6[0].InfoDetail.Org,
		OrgEn:         data.InfoIPv6[0].InfoDetail.OrgEn,
		PostCode:      data.InfoIPv6[0].InfoDetail.PostCode,
		Address:       data.InfoIPv6[0].InfoDetail.Address,
		AddressEn:     data.InfoIPv6[0].InfoDetail.AddressEn,
		NameServer:    strings.Join(data.InfoIPv6[0].InfoDetail.NameServer, ","),
		DsRecord:      data.InfoIPv6[0].InfoDetail.DSRecord,
		NotifyAddress: data.InfoIPv6[0].InfoDetail.NotifyAddress,
		AdminJpnicId:  &[]uint{handles[data.InfoIPv6[0].InfoDetail.AdminJPNICHandle]}[0],
		AsnId:         v6.base.cert.Base.ID,
	})
	if err != nil {
		log.Println("Error", "update result_jpnichandle", err)
		return err
	}
	// JPNIC技術連絡先をDBに登録
	v6TechJPNICLists := []database.V6TechJPNICLists{}
	for _, id := range ids {
		for _, techJPNICHandle := range data.InfoIPv6[0].InfoDetail.TechJPNICHandles {
			v6TechJPNICLists = append(v6TechJPNICLists, database.V6TechJPNICLists{
				V6ListId:      id,
				JpnicHandleId: handles[techJPNICHandle.TechJPNICHandle],
			})
		}
	}
	_, err = v6.base.db.CreateResultV6ListTechJPNIC(v6TechJPNICLists)
	if err != nil {
		log.Println("Error", "prepare(INSERT) result_v4list_tech_jpnic", err)
		return err
	}

	return nil
}
