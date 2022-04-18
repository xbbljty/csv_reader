package main

import (
	"testing"
)

type TestCsv struct {
	Id       int32   `csv:"---"`
	IntVal   int64   `csv:"IntVal"`
	FloatVal float64 `csv:"FloatVal"`
}

func TestDecoder(t *testing.T) {
	fileCsv := `
---,IntVal,FloatVal
Id ,整数     ,浮点数
0  ,234      ,1.05
1  ,345      ,12
2  ,456      ,if error
`
	t.Log(fileCsv)
	d := New(3)
	var dtInfo []*TestCsv
	if err := d.UnMarshalBytes([]byte(fileCsv), &dtInfo); err != nil {
		t.Error(err)
	}

	for index, val := range dtInfo {
		t.Log(index, *val)
	}
}
