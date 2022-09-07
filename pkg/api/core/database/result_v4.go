package database

import (
	"errors"
	"gorm.io/gorm"
)

func (b *Base) CreateResultV4List(handle V4List) (V4List, error) {
	result := b.DB.Table("result_v4list").Create(&handle)
	if result.Error != nil {
		return handle, result.Error
	}

	return handle, nil
}

func (b *Base) GetV4JPNICDataReceivedCount(getStartDate, getEndDate string, AsnID uint) (int64, error) {
	var counter int64
	result := b.DB.Table("result_v4list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND get_start_date < ? AND asn_id = ? AND address != ? AND address_en != ? ",
			getStartDate, getEndDate, AsnID, "", "").Count(&counter)
	if result.Error != nil {
		return counter, result.Error
	}

	return counter, result.Error
}

func (b *Base) GetV4JPNICDataNotReceivedCount(getStartDate, getEndDate string, AsnID uint) (int64, error) {
	var counter int64
	result := b.DB.Table("result_v4list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND get_start_date < ? AND asn_id = ? AND address = ? AND address_en = ? ",
			getStartDate, getEndDate, AsnID, "", "").Count(&counter)
	if result.Error != nil {
		return counter, result.Error
	}

	return counter, result.Error
}

func (b *Base) GetRangeV4List(getStartDate string, AsnID uint) (V4List, error) {
	var jpnicHandleList V4List
	result := b.DB.Table("result_v4list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND asn_id = ? AND address = ? AND address_en = ?", getStartDate, AsnID, "", "").
		First(&jpnicHandleList)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return jpnicHandleList, nil
		}
		return jpnicHandleList, result.Error
	}

	return jpnicHandleList, nil
}

func (b *Base) GetRangeV4ListByRecepNumber(getStartDate, recepNumber string, AsnID uint) ([]V4List, error) {
	var jpnicHandleList []V4List
	result := b.DB.Table("result_v4list").Select("id", "ip_address", "recep_number").
		Where("get_start_date >= ? AND asn_id = ? AND address = ? AND address_en = ? AND recep_number = ?",
			getStartDate, AsnID, "", "", recepNumber).
		First(&jpnicHandleList)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return jpnicHandleList, nil
		}
		return jpnicHandleList, result.Error
	}

	return jpnicHandleList, nil
}

func (b *Base) UpdateV4List(list V4List) error {
	result := b.DB.Table("result_v4list").Model(&V4List{ID: list.ID}).
		Updates(map[string]interface{}{
			"get_date":       list.GetDate,
			"org":            list.Org,
			"org_en":         list.OrgEn,
			"post_code":      list.PostCode,
			"address":        list.Address,
			"address_en":     list.AddressEn,
			"name_server":    list.NameServer,
			"ds_record":      list.DsRecord,
			"notify_address": list.NotifyAddress,
			"admin_jpnic_id": list.AdminJpnicId,
			"asn_id":         list.AsnId,
		})
	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (b *Base) UpdateV4ListIrregular(id uint, getStartDate string, AsnId uint) error {
	result := b.DB.Table("result_v4list").Select(&V4List{ID: id}).
		Updates(map[string]interface{}{
			"get_date":     getStartDate,
			"address":      "　",
			"address_en":   "　",
			"recep_number": "　",
			"asn_id":       AsnId,
		})
	if result.Error != nil {
		return result.Error
	}

	return nil
}
