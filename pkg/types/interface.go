package types

import (
	"reflect"
)

// IsNilInterface returns whether the interface parameter is nil
func IsNilInterface(i interface{}) bool {
	return i == nil || reflect.ValueOf(i).IsNil()
}
