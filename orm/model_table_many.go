package orm

import (
	"fmt"
	"reflect"
)

type manyModel struct {
	*sliceTableModel
	baseTable *Table
	rel       *Relation

	buf       []byte
	dstValues map[string][]reflect.Value
}

var _ TableModel = (*manyModel)(nil)

func newManyModel(j *join) *manyModel {
	baseTable := j.BaseModel.Table()
	joinModel := j.JoinModel.(*sliceTableModel)
	dstValues := dstValues(joinModel, j.Rel.FKValues)
	if len(dstValues) == 0 {
		return nil
	}
	m := manyModel{
		sliceTableModel: joinModel,
		baseTable:       baseTable,
		rel:             j.Rel,

		dstValues: dstValues,
	}
	if !m.sliceOfPtr {
		m.strct = reflect.New(m.table.Type).Elem()
	}
	return &m
}

func (m *manyModel) NextColumnScanner() ColumnScanner {
	if m.sliceOfPtr {
		m.strct = reflect.New(m.table.Type).Elem()
	} else {
		m.strct.Set(m.table.zeroStruct)
	}
	m.structInited = false
	return m
}

func (m *manyModel) AddColumnScanner(model ColumnScanner) error {
	m.buf = modelID(m.buf[:0], m.strct, m.rel.FKs)
	dstValues, ok := m.dstValues[string(m.buf)]
	if !ok {
		return fmt.Errorf(
			"pg: relation=%q has no base model=%q with id=%q (check join conditions)",
			m.rel.Field.GoName, m.baseTable.TypeName, m.buf)
	}

	for _, v := range dstValues {
		if m.sliceOfPtr {
			v.Set(reflect.Append(v, m.strct.Addr()))
		} else {
			v.Set(reflect.Append(v, m.strct))
		}
	}

	return nil
}
