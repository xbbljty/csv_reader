package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
)

type Decoder struct {
	header        map[string]int
	dataStartLine int
}

type CsvMarshal interface {
	FromString(string) error
}

func New(dataStartLine int) *Decoder {
	return &Decoder{dataStartLine: dataStartLine}
}

func (d *Decoder) ResetHeader() {
	d.header = nil
}

func (d *Decoder) WithHeader(header []string) *Decoder {
	d.header = make(map[string]int)
	for i, h := range header {
		d.header[h] = i
	}
	return d
}

func (d *Decoder) UnMarshal(reader *csv.Reader, bean interface{}) error {

	beanT := reflect.TypeOf(bean)
	if beanT.Kind() == reflect.Ptr {
		beanT = beanT.Elem()
	}

	if beanT.Kind() == reflect.Slice {
		beanT = beanT.Elem()
	} else {
		panic("bean is not a slice")
	}
	value := reflect.ValueOf(bean)

	if value.Kind() != reflect.Ptr {
		return fmt.Errorf("bean should be ptr")
	}

	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	} else {
		panic("bean is not a point to slice")
	}
	if d.header == nil {
		row, err := reader.Read()
		if err != nil {
			return err
		}
		d.WithHeader(row)
	}

	count := d.dataStartLine - 2
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		if count > 0 {
			count--
			continue
		}
		beanV, err := d.unMarshal(row, beanT)
		if err != nil {
			return err
		}
		value.Set(reflect.Append(value, beanV))
	}
	return nil
}

func (d *Decoder) UnMarshalBytes(body []byte, bean interface{}) error {
	csvReader := csv.NewReader(bytes.NewReader(body))
	return d.UnMarshal(csvReader, bean)
}

func (d *Decoder) UnMarshalFile(path string, bean interface{}) error {
	csvFile, err := os.Open(path)
	if err != nil {
		return err
	}

	defer func() {
		if err := csvFile.Close(); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err)
		}
	}()
	csvReader := csv.NewReader(csvFile)
	return d.UnMarshal(csvReader, bean)
}

func (d *Decoder) unMarshal(row []string, beanT reflect.Type) (beanR reflect.Value, err error) {
	var beanV reflect.Value
	var isPtr bool
	if beanT.Kind() != reflect.Struct {
		isPtr = true
		beanT = beanT.Elem()
	}
	beanR = reflect.New(beanT)
	beanV = beanR.Elem()
	var value string
	var index int
	var ok bool
	var fileV reflect.Value
	var fileT reflect.StructField
	for i := 0; i < beanV.NumField(); i++ {
		fileT = beanT.Field(i)
		fileV = beanV.FieldByName(fileT.Name)
		if !fileV.CanSet() {
			fmt.Println(fileT.Name)
			continue
		}
		if tag := fileT.Tag.Get("csv"); tag != "" {
			if tag == "-" {
				continue
			}
			index, ok = d.header[tag]
		} else {
			index, ok = d.getIndex(fileT.Name)
		}
		if ok {
			value = row[index]
			value = strings.TrimSpace(value)
			if fileV.Kind() == reflect.Ptr {
				fileV.Set(reflect.New(fileV.Type().Elem()))
				fileV = fileV.Elem()
			}
			if fileV.CanAddr() {
				ptv := fileV.Addr()
				if ptv.CanInterface() {
					if m, ok := ptv.Interface().(CsvMarshal); ok {
						if err = m.FromString(value); err != nil {
							return
						}
						continue
					}
				}
			} else {
				fmt.Println("can not addr", fileT.Name)
			}
			if err = setField(fileV, value); err != nil {
				return
			}
		}
	}
	if !isPtr {
		return beanV, err
	}
	return
}

func (d *Decoder) getIndex(name string) (index int, ok bool) {
	if index, ok = d.header[name]; ok {
		return
	}
	defer func() {
		if ok {
			d.header[name] = index
		}
	}()
	if index, ok = d.header[ToSnake(name, false)]; !ok {
		if index, ok = d.header[ToSnake(name, true)]; !ok {
			if index, ok = d.header[strings.ToLower(name)]; !ok {
				index, ok = d.header[strings.ToUpper(name)]
			}
		}
	}
	return index, ok
}
