package memstats

import (
	"reflect"
	"runtime"
	"strconv"
)

type StringMS struct {
	Name, Value string
}

var ms runtime.MemStats

func init() {
	runtime.ReadMemStats(&ms)
}

func ReadMemStats() {
	runtime.ReadMemStats(&ms)
}

func GetMemStatsForServer() (stringsMS []StringMS) {
	val := reflect.ValueOf(ms)
	stringsMS = make([]StringMS, 0, 27)

	for fieldIdx := 0; fieldIdx < val.NumField(); fieldIdx++ {
		name := val.Type().Field(fieldIdx).Name
		if !isForServer(name) {
			continue
		}

		field := val.Field(fieldIdx)

		val := ""
		if field.Kind() == reflect.Uint64 || field.Kind() == reflect.Uint32 {
			val = strconv.FormatUint(field.Uint(), 10)
		} else if field.Kind() == reflect.Float64 {
			val = strconv.FormatFloat(field.Float(), 'f', -1, 64)
		}

		stringsMS = append(stringsMS, StringMS{name, val})
	}

	return
}

func isForServer(fieldName string) bool {
	switch fieldName {
	case "PauseNs":
		return false
	case "PauseEnd":
		return false
	case "EnableGC":
		return false
	case "DebugGC":
		return false
	case "BySize":
		return false
	default:
		return true
	}
}
