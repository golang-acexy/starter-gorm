package multipledb_test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-gorm/test/model"
	"github.com/lib/pq"
	"testing"
)

var employeeMapper model.EmployeeMapper

func TestPostgresRaw(t *testing.T) {
	var employees []model.Employee
	db := gormstarter.RawPostgresGormDB()
	db.Raw("select * from employee").Scan(&employees)
	fmt.Println(json.ToJson(employees))
}

func TestPostgresSelect(t *testing.T) {
	var employee model.Employee
	fmt.Println(employeeMapper.SelectById(1, &employee))
	fmt.Println(json.ToJson(employee))

	employee = model.Employee{
		LeaderId: pq.Int32Array([]int32{1, 2, 3}),
	}
	fmt.Println(employeeMapper.SelectOneByCond(&employee, &employee))
}
