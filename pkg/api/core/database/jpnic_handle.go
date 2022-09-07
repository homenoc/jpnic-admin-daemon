package database

func (b *Base) CreateJPNICHandle(handle JPNICHandle) (JPNICHandle, error) {
	result := b.DB.Table("result_jpnichandle").Create(&handle)
	if result.Error != nil {
		return handle, result.Error
	}

	return handle, nil
}

func (b *Base) GetRangeJPNICHandle(getStartDate string, AsnID uint, IsIPv6 bool) (*[]JPNICHandle, error) {
	var jpnicHandleLists []JPNICHandle
	result := b.DB.Table("result_jpnichandle").Select("id", "jpnic_handle", "get_date").
		Where("get_start_date >= ? AND asn_id = ? AND is_ipv6 = ?", getStartDate, AsnID, IsIPv6).Scan(&jpnicHandleLists)
	if result.Error != nil {
		return nil, result.Error
	}

	return &jpnicHandleLists, nil
}
