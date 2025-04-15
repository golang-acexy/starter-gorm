package gormstarter

import (
	"errors"
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/reflect"
	"gorm.io/gorm"
	"math"
)

func (b BaseMapper[T]) rawDB() *gorm.DB {
	if b.Tx != nil {
		return b.Tx
	}
	if len(gormDBs) == 1 {
		return gormDBs[defaultDBType]
	}
	if v, flag := any(b.model).(IBaseModelWithDBType); flag {
		return gormDBs[v.DBType()]
	}
	return gormDBs[defaultDBType]
}

func checkResult(rs *gorm.DB, txCheck ...bool) (int64, error) {
	if rs.Error != nil {
		return 0, rs.Error
	}
	if len(txCheck) > 0 && txCheck[0] {
		// 兼容transaction的检查，如果是查询防止未命中数据时触发回滚
		return math.MaxInt64, nil
	}
	return rs.RowsAffected, nil
}

func (b BaseMapper[T]) Gorm() *gorm.DB {
	return b.rawDB().Table(b.model.TableName())
}

func (b BaseMapper[T]) SelectById(id any, result *T) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where("id = ?", id).Scan(result))
}

func (b BaseMapper[T]) SelectByIds(id []interface{}, result *[]*T) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where("id in ?", id).Scan(result))
}

func (b BaseMapper[T]) SelectOneByCond(condition *T, result *T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Scan(result))
}

func (b BaseMapper[T]) SelectOneByMap(condition map[string]any, result *T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Scan(result))
}

func (b BaseMapper[T]) SelectOneByWhere(rawWhereSql string, result *T, args ...interface{}) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Scan(result))
}

func (b BaseMapper[T]) SelectOneByGorm(result *T, rawDb func(*gorm.DB)) (int64, error) {
	var db = b.rawDB().Table(b.model.TableName())
	rawDb(db)
	return checkResult(db.Scan(result))
}

func (b BaseMapper[T]) SelectByCond(condition *T, orderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Scan(result))
}

func (b BaseMapper[T]) SelectByMap(condition map[string]any, orderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Scan(result))
}

func (b BaseMapper[T]) SelectByWhere(rawWhereSql, orderBy string, result *[]*T, args ...interface{}) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Order(orderBy).Scan(result))
}

func (b BaseMapper[T]) SelectByGorm(result *[]*T, rawDb func(*gorm.DB)) (int64, error) {
	var db = b.rawDB().Table(b.model.TableName())
	rawDb(db)
	return checkResult(db.Scan(result))
}

func (b BaseMapper[T]) CountByCond(condition *T) (int64, error) {
	var count int64
	_, err := checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Count(&count))
	return count, err
}

func (b BaseMapper[T]) CountByMap(condition map[string]any) (int64, error) {
	var count int64
	_, err := checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Count(&count))
	return count, err
}

func (b BaseMapper[T]) CountByWhere(rawWhereSql string, args ...interface{}) (int64, error) {
	var count int64
	_, err := checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Count(&count))
	return count, err
}

func (b BaseMapper[T]) CountByGorm(raw func(*gorm.DB)) (int64, error) {
	var count int64
	var db = b.rawDB().Table(b.model.TableName())
	raw(db)
	_, err := checkResult(db.Count(&count))
	return count, err
}

func (b BaseMapper[T]) SelectPageByCond(condition *T, orderBy string, pageNumber, pageSize int, result *[]*T, specifyColumns ...string) (total int64, err error) {
	if pageNumber <= 0 || pageSize <= 0 {
		return 0, errors.New("pageNumber or pageSize <= 0")
	}
	_, err = checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (b BaseMapper[T]) SelectPageByMap(condition map[string]any, orderBy string, pageNumber, pageSize int, result *[]*T, specifyColumns ...string) (total int64, err error) {
	if pageNumber <= 0 || pageSize <= 0 {
		return 0, errors.New("pageNumber or pageSize <= 0")
	}
	_, err = checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (b BaseMapper[T]) SelectPageByWhere(rawWhereSql, orderBy string, pageNumber, pageSize int, result *[]*T, args ...interface{}) (total int64, err error) {
	if pageNumber <= 0 || pageSize <= 0 {
		return 0, errors.New("pageNumber or pageSize <= 0")
	}
	_, err = checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	_, err = checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Order(orderBy).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

func (b BaseMapper[T]) Save(entity *T, excludeColumns ...string) (int64, error) {
	var db = b.rawDB()
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Create(entity))
}

func (b BaseMapper[T]) SaveWithoutZeroField(entity *T) (int64, error) {
	nonZeroFields, err := reflect.NonZeroField(entity)
	if err != nil {
		return 0, err
	}
	if len(nonZeroFields) == 0 {
		return 0, errors.New("no field to save")
	}
	if len(nonZeroFields) == 1 {
		return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields[0]).Create(entity))
	} else {
		nonZeroFieldsSlice := coll.SliceCollect(nonZeroFields[1:], func(t string) interface{} {
			return t
		})
		return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields[0], nonZeroFieldsSlice...).Create(entity))
	}
}

func (b BaseMapper[T]) SaveBatch(entities *[]*T, excludeColumns ...string) (int64, error) {
	var db = b.rawDB()
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Create(entities))
}

func (b BaseMapper[T]) SaveOrUpdateByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	var db = b.rawDB()
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Save(entity))
}

func (b BaseMapper[T]) UpdateById(updated *T, updateColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(updateColumns).Updates(updated))
}

func (b BaseMapper[T]) UpdateByIdWithoutZeroField(updated *T, allowZeroFiledColumns ...string) (int64, error) {
	nonZeroFields, err := reflect.NonZeroField(updated)
	if err != nil {
		return 0, err
	}
	if len(allowZeroFiledColumns) > 0 {
		nonZeroFields = append(nonZeroFields, allowZeroFiledColumns...)
	}
	nonZeroFields = coll.SliceDistinct(nonZeroFields)
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields).Updates(updated))
}

func (b BaseMapper[T]) UpdateByIdUseMap(updated map[string]any, id any) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where("id = ?", id).Updates(updated))
}

func (b BaseMapper[T]) UpdateByCond(updated, condition *T, updateColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(updateColumns).Where(condition).Updates(updated))
}

func (b BaseMapper[T]) UpdateByCondWithZeroField(updated, condition *T, allowZeroFiledColumns []string) (int64, error) {
	nonZeroFields, err := reflect.NonZeroField(updated)
	if err != nil {
		return 0, err
	}
	if len(allowZeroFiledColumns) > 0 {
		nonZeroFields = append(nonZeroFields, allowZeroFiledColumns...)
	}
	nonZeroFields = coll.SliceDistinct(nonZeroFields)
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields).Where(condition).Updates(updated))
}

func (b BaseMapper[T]) UpdateByMap(updated, condition map[string]any) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Updates(updated))
}

func (b BaseMapper[T]) UpdateByWhere(updated *T, rawWhereSql string, args ...interface{}) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Updates(updated))
}

func (b BaseMapper[T]) DeleteById(id ...any) (int64, error) {
	return checkResult(b.rawDB().Delete(b.model, id))
}

func (b BaseMapper[T]) DeleteByCond(condition *T) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Delete(b.model))
}

func (b BaseMapper[T]) DeleteByWhere(rawWhereSql string, args ...interface{}) (int64, error) {
	return checkResult(b.rawDB().Where(rawWhereSql, args...).Delete(b.model))
}

func (b BaseMapper[T]) DeleteByMap(condition map[string]any) (int64, error) {
	return checkResult(b.rawDB().Where(condition).Delete(b.model))
}
