package tripod

import (
	"github.com/sirupsen/logrus"
	"github.com/yu-org/yu/common/yerror"
	"reflect"
	"strings"
	"unsafe"
)

var (
	TripodInjectTag = "tripod"
	BronzeInjectTag = "bronze"
	OmitEmpty       = "omitempty"
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
		field := triStruct.Field(i)

		err := injectTripod(tri, field, tag)
		if err != nil {
			return err
		}

		err = injectBronze(tri, field, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

func injectTripod(tri *Tripod, field reflect.Value, tag reflect.StructTag) error {
	tagValue, hasTag := tag.Lookup(TripodInjectTag)
	if !hasTag {
		return nil
	}
	arr := strings.Split(tagValue, ",")
	tripodName := arr[0]
	triToInject, ok := tri.Land.tripodsMap[tripodName]
	if !ok {
		if len(arr) > 1 {
			if arr[1] == OmitEmpty {
				return nil
			}
		}
		return yerror.TripodNotFound(tripodName)
	}
	if field.CanSet() {
		field.Set(reflect.ValueOf(triToInject.Instance))
	} else {
		// set private field
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		field.Set(reflect.ValueOf(triToInject.Instance))
	}
	logrus.Debugf("inject tripod(%s) into %s", triToInject.Name(), tri.name)

	return nil
}

func injectBronze(tri *Tripod, field reflect.Value, tag reflect.StructTag) error {
	tagValue, hasTag := tag.Lookup(BronzeInjectTag)
	if !hasTag {
		return nil
	}
	arr := strings.Split(tagValue, ",")
	bronzeName := arr[0]
	bronzeToInject, ok := tri.Land.bronzes[bronzeName]
	if !ok {
		if len(arr) > 1 {
			if arr[1] == OmitEmpty {
				return nil
			}
		}
		return yerror.BronzeNotFound(bronzeName)
	}
	if field.CanSet() {
		field.Set(reflect.ValueOf(bronzeToInject.Instance))
	} else {
		// set private field
		field = reflect.NewAt(field.Type(), unsafe.Pointer(field.UnsafeAddr())).Elem()
		field.Set(reflect.ValueOf(bronzeToInject.Instance))
	}
	logrus.Debugf("inject bronze(%s) into %s", bronzeToInject.Name(), tri.name)
	return nil
}
