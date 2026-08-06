package gormstarter

import (
	"errors"
	"testing"
)

type validationModel struct {
	ID   int64
	Name string
}

func (validationModel) TableName() string { return "validation_model" }

func TestWriteValidation(t *testing.T) {
	mapper := BaseMapper[validationModel]{model: validationModel{}}

	tests := []struct {
		name string
		run  func() error
		want error
	}{
		{name: "nil insert", run: func() error { _, err := mapper.Insert(nil); return err }, want: ErrNilEntity},
		{name: "nil non-zero insert", run: func() error { _, err := mapper.InsertWithoutZeroFields(nil); return err }, want: ErrNilEntity},
		{name: "nil upsert", run: func() error { _, err := mapper.InsertOrUpdateByPrimaryKey(nil); return err }, want: ErrNilEntity},
		{name: "empty batch", run: func() error { _, err := mapper.InsertBatch(nil); return err }, want: ErrNoFieldToSave},
		{name: "nil batch entity", run: func() error { _, err := mapper.InsertBatch([]*validationModel{nil}); return err }, want: ErrNilEntity},
		{name: "nil update", run: func() error { _, err := mapper.UpdateByID(nil); return err }, want: ErrNilEntity},
		{name: "empty update", run: func() error { _, err := mapper.UpdateByID(&validationModel{ID: 1}); return err }, want: ErrNoFieldToUpdate},
		{name: "empty non-zero update", run: func() error { _, err := mapper.UpdateByIDWithoutZeroFields(&validationModel{ID: 1}); return err }, want: ErrNoFieldToUpdate},
		{name: "nil condition update", run: func() error { _, err := mapper.UpdateByCond(nil, validationModel{ID: 1}); return err }, want: ErrNilEntity},
		{name: "empty update condition", run: func() error { _, err := mapper.UpdateByCond(&validationModel{Name: "updated"}, validationModel{}); return err }, want: ErrEmptyCondition},
		{name: "empty zero-field update", run: func() error { _, err := mapper.UpdateByCondWithZeroFields(&validationModel{}, validationModel{ID: 1}); return err }, want: ErrNoFieldToUpdate},
		{name: "empty zero-field condition", run: func() error { _, err := mapper.UpdateByCondWithZeroFields(&validationModel{}, validationModel{}, "name"); return err }, want: ErrEmptyCondition},
		{name: "nil update where", run: func() error { _, err := mapper.UpdateByWhere(nil, "id = ?", 1); return err }, want: ErrNilEntity},
		{name: "empty update where", run: func() error { _, err := mapper.UpdateByWhere(&validationModel{Name: "updated"}, " "); return err }, want: ErrEmptyWhereSQL},
		{name: "empty delete IDs", run: func() error { _, err := mapper.DeleteByIDs(nil); return err }, want: ErrEmptyIDs},
		{name: "empty delete condition", run: func() error { _, err := mapper.DeleteByCond(validationModel{}); return err }, want: ErrEmptyCondition},
		{name: "empty delete where", run: func() error { _, err := mapper.DeleteByWhere(" "); return err }, want: ErrEmptyWhereSQL},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.run(); !errors.Is(err, test.want) {
				t.Fatalf("校验错误 = %v，期望 %v", err, test.want)
			}
		})
	}
}
