package tripod

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common/yerror"
	"reflect"
	"strings"
	"unsafe"
)

var (
	InjectTag = "tripod"
	OmitEmpty = "omitempty"
)

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
		tagValue, hasTag := tag.Lookup(InjectTag)
		if !hasTag {
			continue
		}
		arr := strings.Split(tagValue, ",")
		tripodName := arr[0]
		triToInject, ok := tri.Land.TripodsMap[tripodName]
		if !ok {
			if len(arr) > 1 {
				if arr[1] == OmitEmpty {
					continue
				}
			}
			return yerror.TripodNotFound(tripodName)
		}
		field := triStruct.Field(i)
		if field.CanSet() {
			field.Set(reflect.ValueOf(triToInject.Instance))
		} else {
			// set private field
			field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
			field.Set(reflect.ValueOf(triToInject.Instance))
		}
		logrus.Debugf("inject tripod(%s) into %s", tripodName, tri.name)
	}
	return nil
}
