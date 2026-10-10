package message

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrNotStructPtr = errors.New("bind: target must be a pointer to a struct")
	ErrNotStruct    = errors.New("bind: source must be a struct or struct pointer")
)

// Marshal converts a Go struct annotated with `iso:"..."` tags into a Message.
// Tag syntax:
//
//	`iso:"0"`         -> field 0 (MTI)
//	`iso:"2"`         -> field 2
//	`iso:"48.1"`      -> subfield 1 of field 48
func Marshal(v any) (*Message, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	msg := New()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		fieldStruct := typ.Field(i)
		if fieldStruct.PkgPath != "" {
			continue
		}
		tag := fieldStruct.Tag.Get("iso")
		if tag == "" || tag == "-" {
			continue
		}

		fieldVal := val.Field(i)
		if !fieldVal.IsValid() {
			continue
		}

		if strings.Contains(tag, ".") {
			// Subfield path (e.g. "48.1")
			strVal := formatVal(fieldVal)
			if strVal != "" {
				if err := msg.SetSubfieldByPath(tag, strVal); err != nil {
					return nil, err
				}
			}
			continue
		}

		fieldID, err := strconv.Atoi(tag)
		if err != nil {
			return nil, fmt.Errorf("bind: invalid iso tag '%s': %w", tag, err)
		}

		switch fieldVal.Kind() {
		case reflect.Slice:
			if fieldVal.Type().Elem().Kind() == reflect.Uint8 {
				b := fieldVal.Bytes()
				if len(b) > 0 {
					msg.SetBytes(fieldID, b)
				}
			}
		default:
			strVal := formatVal(fieldVal)
			if strVal != "" {
				msg.Set(fieldID, strVal)
			}
		}
	}

	return msg, nil
}

// Unmarshal populates a struct pointer with values from an ISO-8583 Message based on `iso:"..."` tags.
func Unmarshal(msg *Message, v any) error {
	if msg == nil {
		return errors.New("bind: nil message")
	}
	val := reflect.ValueOf(v)
	if val.Kind() != reflect.Pointer || val.Elem().Kind() != reflect.Struct {
		return ErrNotStructPtr
	}

	val = val.Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		fieldStruct := typ.Field(i)
		if fieldStruct.PkgPath != "" {
			continue
		}
		tag := fieldStruct.Tag.Get("iso")
		if tag == "" || tag == "-" {
			continue
		}

		targetField := val.Field(i)
		if !targetField.CanSet() {
			continue
		}

		if strings.Contains(tag, ".") {
			subVal, ok := msg.GetSubfield(tag)
			if ok {
				if err := setVal(targetField, subVal); err != nil {
					return err
				}
			}
			continue
		}

		fieldID, err := strconv.Atoi(tag)
		if err != nil {
			return fmt.Errorf("bind: invalid iso tag '%s': %w", tag, err)
		}

		comp, ok := msg.Get(fieldID)
		if !ok {
			continue
		}

		if targetField.Kind() == reflect.Slice && targetField.Type().Elem().Kind() == reflect.Uint8 {
			targetField.SetBytes(comp.GetBytes())
			continue
		}

		if err := setVal(targetField, comp.GetString()); err != nil {
			return err
		}
	}

	return nil
}

func formatVal(v reflect.Value) string {
	switch v.Kind() {
	case reflect.String:
		return v.String()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(v.Uint(), 10)
	default:
		return fmt.Sprintf("%v", v.Interface())
	}
}

func setVal(target reflect.Value, strVal string) error {
	switch target.Kind() {
	case reflect.String:
		target.SetString(strVal)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(strVal, 10, target.Type().Bits())
		if err != nil {
			clean := *err.(*strconv.NumError)
			clean.Num = ""
			return fmt.Errorf("bind: invalid signed integer: %w", &clean)
		}
		target.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(strVal, 10, target.Type().Bits())
		if err != nil {
			clean := *err.(*strconv.NumError)
			clean.Num = ""
			return fmt.Errorf("bind: invalid unsigned integer: %w", &clean)
		}
		target.SetUint(uintVal)
	default:
		return fmt.Errorf("bind: unsupported struct field kind %s", target.Kind())
	}
	return nil
}
