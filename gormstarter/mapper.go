package gormstarter

import (
	"database/sql"
	"errors"
	"github.com/acexy/golang-toolkit/util/coll"
	"github.com/acexy/golang-toolkit/util/reflect"
	"gorm.io/gorm"
	"math"
)

func (b BaseMapper[T]) rawDB() *gorm.DB {
	if b.tx != nil {
		return b.tx
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

// GormWithTableName Mapper对应的原生Gorm操作能力 获取到的原始gorm.DB已经限定当前Mapper所对应的表名
func (b BaseMapper[T]) GormWithTableName() *gorm.DB {
	return b.rawDB().Table(b.model.TableName())
}

// CurrentGorm 获取当前Mapper所使用的gorm.DB 如果当前Mapper已使用指定的事务，则返回当前Mapper所使用的事务，否则获取新的gorm.DB
func (b BaseMapper[T]) CurrentGorm() *gorm.DB {
	if b.tx != nil {
		return b.tx
	}
	return b.rawDB()
}

// GetBaseMapperWithTx 获取携带指定事务的基础Mapper
func (b BaseMapper[T]) GetBaseMapperWithTx(tx *gorm.DB) BaseMapper[T] {
	return BaseMapper[T]{
		model: b.model,
		tx:    tx,
	}
}

// NewBaseMapperWithTx 创建一个全新事务的基础Mapper
func (b BaseMapper[T]) NewBaseMapperWithTx(opts ...*sql.TxOptions) BaseMapper[T] {
	baseMapper := BaseMapper[T]{
		model: b.model,
	}
	baseMapper.tx = baseMapper.rawDB().Begin(opts...)
	return baseMapper
}

// SelectById 通过主键查询数据
func (b BaseMapper[T]) SelectById(id any, result *T) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where("id = ?", id).Scan(result))
}

// SelectByIds 通过主键查询数据
func (b BaseMapper[T]) SelectByIds(id []any, result *[]*T) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where("id in ?", id).Scan(result))
}

// SelectOneByCond 通过条件查询 查询条件零值字段将被自动忽略
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectOneByCond(condition, result *T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Scan(result))
}

// SelectOneByMap 通过指定字段与值查询数据 解决查询条件零值问题
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectOneByMap(condition map[string]any, result *T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Scan(result))
}

// SelectOneByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b BaseMapper[T]) SelectOneByWhere(rawWhereSql string, result *T, args ...any) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Scan(result))
}

// SelectOneByGorm 通过原始Gorm查询单条数据 构建Gorm查询条件
func (b BaseMapper[T]) SelectOneByGorm(result *T, rawDb func(*gorm.DB)) (int64, error) {
	var db = b.rawDB().Table(b.model.TableName())
	rawDb(db)
	return checkResult(db.Scan(result))
}

// SelectByCond 通过条件查询 查询条件零值字段将被自动忽略
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectByCond(condition *T, orderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Scan(result))
}

// SelectByMap 通过指定字段与值查询数据 解决零值条件问题
// specifyColumns 指定只需要查询的数据库字段
func (b BaseMapper[T]) SelectByMap(condition map[string]any, orderBy string, result *[]*T, specifyColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(condition).Order(orderBy).Scan(result))
}

// SelectByWhere 通过原始Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b BaseMapper[T]) SelectByWhere(rawWhereSql, orderBy string, result *[]*T, args ...any) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Order(orderBy).Scan(result))
}

// SelectByGorm 通过原始Gorm查询数据
func (b BaseMapper[T]) SelectByGorm(result *[]*T, rawDb func(*gorm.DB)) (int64, error) {
	var db = b.rawDB().Table(b.model.TableName())
	rawDb(db)
	return checkResult(db.Scan(result))
}

// CountByCond 通过条件查询数据总数 查询条件零值字段将被自动忽略
func (b BaseMapper[T]) CountByCond(condition *T) (int64, error) {
	var count int64
	_, err := checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Count(&count))
	return count, err
}

// CountByMap 通过指定字段与值查询数据总数 解决零值条件问题
func (b BaseMapper[T]) CountByMap(condition map[string]any) (int64, error) {
	var count int64
	_, err := checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Count(&count))
	return count, err
}

// CountByWhere 通过原始SQL查询数据总数
func (b BaseMapper[T]) CountByWhere(rawWhereSql string, args ...any) (int64, error) {
	var count int64
	_, err := checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Count(&count))
	return count, err
}

// CountByGorm 通过原始Gorm查询数据总数
func (b BaseMapper[T]) CountByGorm(raw func(*gorm.DB)) (int64, error) {
	var count int64
	var db = b.rawDB().Table(b.model.TableName())
	raw(db)
	_, err := checkResult(db.Count(&count))
	return count, err
}

// SelectPageByCond 通过条件分页查询 零值字段将被自动忽略
// specifyColumns 指定只需要查询的数据库字段 pageNumber 页码 1开始
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

// SelectPageByMap 通过指定字段与值查询数据分页查询 解决零值条件问题
// specifyColumns 指定只需要查询的数据库字段 pageNumber 页码 1开始
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

// SelectPageByWhere 通过原始SQL分页查询 rawWhereSql 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b BaseMapper[T]) SelectPageByWhere(rawWhereSql, orderBy string, pageNumber, pageSize int, result *[]*T, args []any, specifyColumns ...string) (total int64, err error) {
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
	_, err = checkResult(b.rawDB().Table(b.model.TableName()).Select(specifyColumns).Where(rawWhereSql, args...).Order(orderBy).Limit(pageSize).Offset((pageNumber - 1) * pageSize).Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// SelectPageByGorm 通过原始Gorm分页查询
func (b BaseMapper[T]) SelectPageByGorm(countRawDb func(*gorm.DB), pageRawDb func(*gorm.DB), result *[]*T) (total int64, err error) {
	var countDb = b.rawDB().Table(b.model.TableName())
	countRawDb(countDb)
	_, err = checkResult(countDb.Count(&total))
	if err != nil {
		return 0, err
	}
	if total <= 0 {
		return 0, nil
	}
	selectDb := b.rawDB().Table(b.model.TableName())
	pageRawDb(selectDb)
	_, err = checkResult(selectDb.Scan(result))
	if err != nil {
		return 0, err
	}
	return total, nil
}

// Insert 保存数据 零值也将参与保存
//
//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段名称
func (b BaseMapper[T]) Insert(entity *T, excludeColumns ...string) (int64, error) {
	var db = b.rawDB()
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Create(entity))
}

// InsertWithoutZeroField 保存数据 零值将不会参与保存
func (b BaseMapper[T]) InsertWithoutZeroField(entity *T) (int64, error) {
	nonZeroFields, err := reflect.NonZeroFieldName(entity)
	if err != nil {
		return 0, err
	}
	if len(nonZeroFields) == 0 {
		return 0, errors.New("no field to save")
	}
	if len(nonZeroFields) == 1 {
		return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields[0]).Create(entity))
	} else {
		nonZeroFieldsSlice := coll.SliceCollect(nonZeroFields[1:], func(t string) any {
			return t
		})
		return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields[0], nonZeroFieldsSlice...).Create(entity))
	}
}

// InsertBatch 批量新增 零值也将参与保存
//
//	exclude 手动指定需要排除的字段名称 数据库字段/结构体字段
func (b BaseMapper[T]) InsertBatch(entities *[]*T, excludeColumns ...string) (int64, error) {
	var db = b.rawDB()
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Create(entities))
}

// InsertUseMap 通过Map类型保存数据
func (b BaseMapper[T]) InsertUseMap(entity map[string]any) (int64, error) {
	if len(entity) == 0 {
		return 0, errors.New("no field to save")
	}
	return checkResult(b.rawDB().Create(entity))
}

// InsertOrUpdateByPrimaryKey 保存/更新数据 零值也将参与保存
// exclude 手动指定需要排除的字段名称 数据库字段/结构体字段 (如果触发的是update 创建时间可能会被错误的修改，可以通过excludeColumns来指定排除创建时间字段)
// 仅根据主键冲突默认支持update 更多操作需要参阅 https://gorm.io/zh_CN/docs/create.html#upsert
func (b BaseMapper[T]) InsertOrUpdateByPrimaryKey(entity *T, excludeColumns ...string) (int64, error) {
	var db = b.rawDB()
	if len(excludeColumns) > 0 {
		db = db.Omit(excludeColumns...)
	}
	return checkResult(db.Save(entity))
}

// UpdateById 通过ID更新含零值字段
// updateColumns 手动指定需要更新的列
func (b BaseMapper[T]) UpdateById(updated *T, updateColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(updateColumns).Updates(updated))
}

// UpdateByIdWithoutZeroField 通过ID更新非零值字段
// allowZeroFiledColumns 额外指定需要更新零值字段
func (b BaseMapper[T]) UpdateByIdWithoutZeroField(updated *T, allowZeroFiledColumns ...string) (int64, error) {
	nonZeroFields, err := reflect.NonZeroFieldName(updated)
	if err != nil {
		return 0, err
	}
	if len(allowZeroFiledColumns) > 0 {
		nonZeroFields = append(nonZeroFields, allowZeroFiledColumns...)
	}
	nonZeroFields = coll.SliceDistinct(nonZeroFields)
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields).Updates(updated))
}

// UpdateByIdUseMap 通过ID更新所有map中指定的列和值
func (b BaseMapper[T]) UpdateByIdUseMap(updated map[string]any, id any) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where("id = ?", id).Updates(updated))
}

// UpdateByCond 通过条件更新 条件：零值将自动忽略，更新：零值字段将被自动忽略
// updateColumns 需要指定更新的数据库字段 更新指定字段(支持零值字段)
func (b BaseMapper[T]) UpdateByCond(updated, condition *T, updateColumns ...string) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(updateColumns).Where(condition).Updates(updated))
}

// UpdateByCondWithZeroField 通过条件更新，并指定可以更新的零值字段
func (b BaseMapper[T]) UpdateByCondWithZeroField(updated, condition *T, allowZeroFiledColumns []string) (int64, error) {
	nonZeroFields, err := reflect.NonZeroFieldName(updated)
	if err != nil {
		return 0, err
	}
	if len(allowZeroFiledColumns) > 0 {
		nonZeroFields = append(nonZeroFields, allowZeroFiledColumns...)
	}
	nonZeroFields = coll.SliceDistinct(nonZeroFields)
	return checkResult(b.rawDB().Table(b.model.TableName()).Select(nonZeroFields).Where(condition).Updates(updated))
}

// UpdateByMap 通过Map类型条件更新
func (b BaseMapper[T]) UpdateByMap(updated, condition map[string]any) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Updates(updated))
}

// UpdateByWhere 通过原始SQL查询条件，更新非零实体字段 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b BaseMapper[T]) UpdateByWhere(updated *T, rawWhereSql string, args ...any) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(rawWhereSql, args...).Updates(updated))
}

// DeleteById 通过ID删除相关数据
func (b BaseMapper[T]) DeleteById(id ...any) (int64, error) {
	return checkResult(b.rawDB().Delete(b.model, id))
}

// DeleteByCond 通过条件删除 零值字段将被自动忽略
func (b BaseMapper[T]) DeleteByCond(condition *T) (int64, error) {
	return checkResult(b.rawDB().Table(b.model.TableName()).Where(condition).Delete(b.model))
}

// DeleteByWhere 通过原始SQL删除相关数据 Where SQL查询 只需要输入SQL语句和参数 例如 where a = 1 则只需要rawWhereSql = "a = ?" args = 1
func (b BaseMapper[T]) DeleteByWhere(rawWhereSql string, args ...any) (int64, error) {
	return checkResult(b.rawDB().Where(rawWhereSql, args...).Delete(b.model))
}

// DeleteByMap 通过Map类型条件删除
func (b BaseMapper[T]) DeleteByMap(condition map[string]any) (int64, error) {
	return checkResult(b.rawDB().Where(condition).Delete(b.model))
}
