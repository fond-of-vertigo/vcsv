package vcsv

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type fieldContext struct {
	structField reflect.StructField
	rv          reflect.Value
	tagOpts     tagOptions
}

func (r *CSVReader) parseFieldValue(structField reflect.StructField, rv reflect.Value) error {
	tagOpts, err := readTag(structField.Tag)
	if err != nil {
		return err
	}
	if tagOpts == nil {
		return nil
	}
	fc := fieldContext{structField: structField, rv: rv, tagOpts: *tagOpts}
	return r.handleFieldByTagOptions(fc)
}

func (r *CSVReader) handleFieldByTagOptions(fc fieldContext) error {
	if fc.tagOpts.columnName != "" {
		return r.handleFieldByName(fc.tagOpts.columnName, fc)
	}

	if fc.tagOpts.index >= 0 {
		return r.handleFieldByIndex(fc.tagOpts.index, fc)
	}

	return nil
}

func (r *CSVReader) handleFieldByName(columnName string, fc fieldContext) error {
	value, err := r.Get(columnName)
	if err != nil {
		return err
	}

	return r.setFieldValue(value, fc)
}

func (r *CSVReader) handleFieldByIndex(index int, fc fieldContext) error {
	if index >= len(r.columns) {
		return fmt.Errorf("index %d out of range for field %s [%s]", index, fc.structField.Name, fc.structField.Tag)
	}

	value := r.columns[index]
	return r.setFieldValue(value, fc)
}

func (r *CSVReader) setFieldValue(value string, fc fieldContext) error {
	convertedValue, err := convertToType(value, fc.structField, fc.tagOpts)
	if err != nil {
		return fmt.Errorf("failed to convert value %s to type %s in field %s [%s]: %w",
			value, fc.structField.Type.Kind(), fc.structField.Name, fc.structField.Tag, err)
	}

	fc.rv.Set(convertedValue)
	return nil
}

type tagOptions struct {
	columnName string
	index      int
	format     string
}

func readTag(tag reflect.StructTag) (*tagOptions, error) {
	tv := tag.Get("csv")
	if tv == "" || tv == "-" {
		return nil, nil
	}

	t := &tagOptions{index: -1} // declare index as -1 to indicate that it was not set
	for _, opt := range strings.Split(tv, ",") {
		if err := parseTagOption(opt, t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

func parseTagOption(opt string, tag *tagOptions) (err error) {
	opt = strings.TrimSpace(opt)

	switch {
	case strings.HasPrefix(opt, "index:"):
		tag.index, err = parseIndex(opt)
		if err != nil {
			return err
		}
	case strings.HasPrefix(opt, "format:"):
		tag.format = parseFormat(opt)
	default:
		tag.columnName = opt
	}

	return nil
}

func parseIndex(opt string) (int, error) {
	opt = opt[6:] // remove `index:`
	return strconv.Atoi(opt)
}

func parseFormat(opt string) string {
	return opt[7:] // remove `format:`
}
