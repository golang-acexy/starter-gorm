package test

import (
	"testing"

	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-gorm/test/model"
	"github.com/lib/pq"
)

var employeeMapper model.EmployeeMapper

func TestInsert(t *testing.T) {
	save := &model.Employee{
		Name:     "法外狂徒",
		LeaderID: pq.Int32Array([]int32{1, 2, 3}),
	}
	if _, err := employeeMapper.InsertWithoutZeroFields(save); err != nil {
		t.Fatal(err)
	}
	if save.ID == 0 {
		t.Fatal("expected generated employee ID")
	}
}

func TestSelectByIDAndCond(t *testing.T) {
	var employee model.Employee
	if _, err := employeeMapper.SelectByID(1, &employee); err != nil {
		t.Fatal(err)
	}

	employee = model.Employee{
		LeaderID: []int32{1, 2, 3},
	}
	if _, err := employeeMapper.SelectOneByCond(gormstarter.CondQuery[model.Employee]{Condition: employee}, &employee); err != nil {
		t.Fatal(err)
	}
}

func TestExistsByID(t *testing.T) {
	employee := &model.Employee{
		Name:     "exists",
		LeaderID: pq.Int32Array([]int32{1}),
	}
	if _, err := employeeMapper.Insert(employee); err != nil {
		t.Fatal(err)
	}
	defer employeeMapper.DeleteByID(employee.ID)

	exists, err := employeeMapper.ExistsByID(employee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Fatal("expected inserted employee to exist")
	}

	if _, err = employeeMapper.DeleteByID(employee.ID); err != nil {
		t.Fatal(err)
	}
	exists, err = employeeMapper.ExistsByID(employee.ID)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("expected deleted employee not to exist")
	}
}
