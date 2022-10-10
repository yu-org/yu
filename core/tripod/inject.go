package tripod

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common/yerror"
	"reflect"
	"unsafe"
)

var InjectTag = "tripod"

func ResolveTripod(v interface{}) *Tripod {
	tri := reflect.Indirect(reflect.ValueOf(v)).FieldByName("Tripod")
	trip := tri.Interface().(*Tripod)
	return trip
}

func Inject(tripodInterface interface{}) error {
	tri := ResolveTripod(tripodInterface)
	triStruct := reflect.Indirect(reflect.ValueOf(tripodInterface))
	for i := 0; i < triStruct.NumField(); i++ {
		tag := triStruct.Type().Field(i).Tag
		name, hasTag := tag.Lookup(InjectTag)
		if !hasTag {
			continue
		}
		triToInject, ok := tri.Land.TripodsMap[name]
		if !ok {
			return yerror.TripodNotFound(name)
		}
		field := triStruct.Field(i)
		if field.CanSet() {
			field.Set(reflect.ValueOf(triToInject.Instance))
		} else {
			// set private field
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			field.Set(reflect.ValueOf(triToInject.Instance))
		}
		logrus.Debugf("inject tripod(%s) into %s", name, tri.name)
	}
	return nil
}
