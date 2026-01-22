package multipledb_test

import (
	"fmt"
	"testing"

	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-gorm/test/model"
)

var employeeMapper model.EmployeeMapper

func TestPostgresRaw(t *testing.T) {
	var employees []model.Employee
	db := gormstarter.RawPostgresGormDB()
	db.Raw("select * from employee").Scan(&employees)
	fmt.Println(json.ToString(employees))
}

func TestPostgresSelect(t *testing.T) {
	var employee model.Employee
	fmt.Println(employeeMapper.SelectById(1, &employee))
	fmt.Println(json.ToString(employee))

	employee = model.Employee{
		LeaderId: []int32{1, 2, 3},
	}
	fmt.Println(employeeMapper.SelectOneByCond(&employee, &employee))
}
