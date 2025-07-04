package test

import (
	"fmt"
	"github.com/acexy/golang-toolkit/util/json"
	"github.com/golang-acexy/starter-gorm/test/model"
	"github.com/lib/pq"
	"testing"
)

func init() {
	_ = starterLoader.Start()
}

var employeeMapper model.EmployeeMapper

func TestSave(t *testing.T) {
	save := &model.Employee{
		Name:     "法外狂徒",
		LeaderId: pq.Int32Array([]int32{1, 2, 3}),
	}
	fmt.Println(employeeMapper.InsertWithoutZeroField(save))
	fmt.Println(save.ID)
}

func TestSelect(t *testing.T) {
	var employee model.Employee
	fmt.Println(employeeMapper.SelectById(1, &employee))
	fmt.Println(json.ToJson(employee))

	employee = model.Employee{
		LeaderId: pq.Int32Array([]int32{1, 2, 3}),
	}
	fmt.Println(employeeMapper.SelectOneByCond(&employee, &employee))
}
