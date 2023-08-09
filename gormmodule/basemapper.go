package gormmodule

type BaseMapper struct {
}

func (m BaseMapper) Save(entity any) (int64, error) {
	rs := db.Begin().Save(&entity)
	if rs.Error != nil {
		return 0, rs.Error
	}
	return rs.RowsAffected, nil
}
