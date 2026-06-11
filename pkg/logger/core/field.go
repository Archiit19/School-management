package core

import (
	"fmt"
	"time"
)

// Field is a structured log attribute.
type Field struct {
	Key string
	Val any
}

func String(key, val string) Field   { return Field{Key: key, Val: val} }
func Int(key string, val int) Field  { return Field{Key: key, Val: val} }
func Int64(key string, val int64) Field { return Field{Key: key, Val: val} }
func Bool(key string, val bool) Field { return Field{Key: key, Val: val} }
func Any(key string, val any) Field  { return Field{Key: key, Val: val} }
func Err(err error) Field            { return Field{Key: "error", Val: err} }
func Duration(key string, val time.Duration) Field {
	return Field{Key: key, Val: val}
}
func Time(key string, val time.Time) Field { return Field{Key: key, Val: val} }

// KV builds fields from alternating key/value pairs: KV("a", 1, "b", "x").
func KV(pairs ...any) []Field {
	if len(pairs) == 0 {
		return nil
	}
	if len(pairs)%2 != 0 {
		panic(fmt.Sprintf("logger.KV: odd number of arguments (%d)", len(pairs)))
	}
	fields := make([]Field, 0, len(pairs)/2)
	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			key = fmt.Sprint(pairs[i])
		}
		fields = append(fields, Field{Key: key, Val: pairs[i+1]})
	}
	return fields
}
