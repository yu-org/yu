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

func ResolveBronze(v any) *Bronze {
	bro := reflect.Indirect(reflect.ValueOf(v)).FieldByName("Bronze")
	bronze := bro.Interface().(*Bronze)
	return bronze
}

// InjectToTripod injects tripod/bronze into tripod.
func InjectToTripod(tripodInstance any) error {
	tri := ResolveTripod(tripodInstance)
	triStruct := reflect.Indirect(reflect.ValueOf(tripodInstance))
	for i := 0; i < triStruct.NumField(); i++ {
		tag := triStruct.Type().Field(i).Tag
		fieldToBeInjected := triStruct.Field(i)

		err := injectTripodToTripod(tri, fieldToBeInjected, tag)
		if err != nil {
			return err
		}

		err = injectBronzeToTripod(tri, fieldToBeInjected, tag)
		if err != nil {
			return err
		}
	}
	return nil
}

// InjectToBronze injects bronze into bronze.
func InjectToBronze(land *Land, bronzeInstance any) error {
	bro := ResolveBronze(bronzeInstance)
	broStruct := reflect.Indirect(reflect.ValueOf(bronzeInstance))
	for i := 0; i < broStruct.NumField(); i++ {
		tag := broStruct.Type().Field(i).Tag
		fieldToBeInjected := broStruct.Field(i)

		tagValue, hasTag := tag.Lookup(BronzeInjectTag)
		if !hasTag {
			return nil
		}
		arr := strings.Split(tagValue, ",")
		bronzeName := arr[0]
		bronzeToInject, ok := land.bronzes[bronzeName]
		if !ok {
			if len(arr) > 1 {
				if arr[1] == OmitEmpty {
					return nil
				}
			}
			return yerror.BronzeNotFound(bronzeName)
		}
		if fieldToBeInjected.CanSet() {
			fieldToBeInjected.Set(reflect.ValueOf(bronzeToInject.Instance))
		} else {
			// set private field
			fieldToBeInjected = reflect.NewAt(fieldToBeInjected.Type(), unsafe.Pointer(fieldToBeInjected.UnsafeAddr())).Elem()
			fieldToBeInjected.Set(reflect.ValueOf(bronzeToInject.Instance))
		}
		logrus.Debugf("inject bronze(%s) into bronze(%s)", bronzeToInject.Name(), bro.name)
	}
	return nil
}

func injectTripodToTripod(tri *Tripod, fieldToBeInjected reflect.Value, tag reflect.StructTag) error {
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
	if fieldToBeInjected.CanSet() {
		fieldToBeInjected.Set(reflect.ValueOf(triToInject.Instance))
	} else {
		// set private field
		fieldToBeInjected = reflect.NewAt(fieldToBeInjected.Type(), unsafe.Pointer(fieldToBeInjected.UnsafeAddr())).Elem()
		fieldToBeInjected.Set(reflect.ValueOf(triToInject.Instance))
	}
	logrus.Debugf("inject tripod(%s) into tripod(%s)", triToInject.Name(), tri.name)

	return nil
}

func injectBronzeToTripod(tri *Tripod, fieldToBeInjected reflect.Value, tag reflect.StructTag) error {
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
	if fieldToBeInjected.CanSet() {
		fieldToBeInjected.Set(reflect.ValueOf(bronzeToInject.Instance))
	} else {
		// set private field
		fieldToBeInjected = reflect.NewAt(fieldToBeInjected.Type(), unsafe.Pointer(fieldToBeInjected.UnsafeAddr())).Elem()
		fieldToBeInjected.Set(reflect.ValueOf(bronzeToInject.Instance))
	}
	logrus.Debugf("inject bronze(%s) into tripod(%s)", bronzeToInject.Name(), tri.name)
	return nil
}
