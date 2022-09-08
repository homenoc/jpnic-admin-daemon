package database

func (b *Base) CreateResultV6ListTechJPNIC(handle []V6TechJPNICLists) ([]V6TechJPNICLists, error) {
	result := b.DB.Table("result_v6list_tech_jpnic").Create(&handle)
	if result.Error != nil {
		return handle, result.Error
	}

	return handle, nil
}
