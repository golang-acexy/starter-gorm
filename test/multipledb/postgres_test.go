package multipledb

import (
	"os"
	"testing"
	"time"

	"github.com/golang-acexy/starter-gorm/gormstarter"
	"github.com/golang-acexy/starter-gorm/test/model"
)

var employeeMapper model.EmployeeMapper

func TestMain(m *testing.M) {
	if err := starterLoader.Start(); err != nil {
		os.Exit(1)
	}
	code := m.Run()
	if _, err := starterLoader.StopAllByRegisteredOrder(10 * time.Second); err != nil {
		os.Exit(1)
	}
	os.Exit(code)
}

func TestPostgresRaw(t *testing.T) {
	var employees []model.Employee
	db := gormstarter.RawPostgresGormDB()
	if db == nil {
		t.Fatal("postgres database is not initialized")
	}
	if err := db.Raw("select * from employee").Scan(&employees).Error; err != nil {
		t.Fatal(err)
	}
}

func TestPostgresSelect(t *testing.T) {
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
