# starter-gorm

基于`github.com/go-gorm/gorm`封装的数据库服务组件

> 特色功能

- BaseMapper 基础数据库能力快速接入

现在你只要定义好你的数据库实体映射后，立马就可以获得该表的CRUD能力

```go
// Teacher 继承BaseModel 并实现 IBaseModel
type Teacher struct {
	gormstarter.BaseModel[uint64] // 继承基础模型 声明主键类型
	CreatedAt time.Time `gorm:"column:create_time" gorm:"<-:create" json:"createTime"`
	UpdatedAt time.Time `gorm:"column:update_time" gorm:"<-:update" json:"updateTime"`
	Name      string
	Sex       uint
	Age       uint
}

func (Teacher) TableName() string { // 实现gorm接口 注册表名
	return "demo_teacher"
}

// TeacherMapper 声明Teacher 获取基于BaseMapper的能力
type TeacherMapper struct {
	gormstarter.BaseMapper[Teacher]
}

```
无须额外代码，即可使用基础CURD能力

```go
func TestModifyMapById(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.ModifyMapById(132, map[string]any{"name": "Miss A", "sex": 0}))
}

func TestModifyByWhere(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.ModifyByWhere(Teacher{Name: "Alex"}, "name = ? and age > ?", "mapper", 5))
}

func TestRemoveById(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.RemoveById(1))
}

func TestRemoveByWhere(t *testing.T) {
	bm := TeacherMapper{}
	fmt.Println(bm.RemoveByWhere("name = ? and age > ?", "Alex", 5))
}

func TestModifyByCondition(t *testing.T) {
	bm := TeacherMapper{}
	updated := Teacher{Name: "1", Age: 12}
	condition := Teacher{Name: "2", Age: 1}
	fmt.Println(bm.ModifyByCondition(updated, condition))
}

func TestQueryById(t *testing.T) {
	bm := TeacherMapper{}
	var teacher Teacher
	fmt.Println(bm.QueryById(1, &teacher))
	fmt.Println(teacher)
}

func TestQueryByCondition(t *testing.T) {
	bm := TeacherMapper{}
	teachers := new([]*Teacher)
	bm.QueryByCondition(Teacher{Name: "王五", Sex: 1}, teachers)
	fmt.Println(teachers)
}
...
```

- Transaction Auto Commit/Rollback 更友好事务支持

允许在代码中主动开启事务，并在执行异常后主动回滚整个事务链

```go
func TestTransactionPrepareRollback(t *testing.T) {
	tx := gormstarter.NewTransactionPrepare(true)
	tx.Save(&Student{Name: "张三"})
	tx.Save(&Student{Name: "李四"})
	tx.Rollback()
	tx.Save(&Teacher{Name: "王五"})
	fmt.Println(tx.Execute())
}

func TestTransactionRollback(t *testing.T) {
	tx := gormstarter.NewTransaction(true)
	tx.Save(&Student{Name: "张三"})
	tx.Save(&Student{Name: "李四"})
	tx.Rollback()
	tx.Save(&Teacher{Name: "王五"})
	fmt.Println(tx.Execute())
}
```