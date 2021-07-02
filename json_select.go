package json_select

import (
	"errors"
	"fmt"
	"strconv"
)


// Selecter is wrapper around an interface{} with methods for querying
// its contents
type Selecter struct {
	V interface{}
}

// Select returns a Selecter for the query
func (j Selecter) Select(sels ...interface{}) (Selecter, error) {
	v, err := Select(j.V, sels...)
	return Selecter{V: v}, err
}

// SelectBool is like Select but attempts to coerce the selection into a bool.
// An error is returned if the coercion fails
func (j Selecter) SelectBool(sels ...interface{}) (bool, error) {
	v, err := Select(j.V, sels...)
	if err != nil {
		return false, err
	}

	boolean, ok := v.(bool)
	if !ok {
		return false, fmt.Errorf("%v not a bool", v)
	}

	return boolean, nil
}

// SelectInt is like Select but attempts to coerce and convert the selection
// into an int. An error is returned if the coercion or conversion fails. The
// followin types are supported:
//		int
//		float64
//		string (using strconv.Atoi)
func (j Selecter) SelectInt(sels ...interface{}) (int, error) {
	v, err := Select(j.V, sels...)
	if err != nil {
		return 0, err
	}

	switch vv := v.(type) {
	case int:
		return vv, nil
	case float64:
		return int(vv), nil
	case string:
		i, err := strconv.Atoi(vv)
		if err != nil {
			return 0, fmt.Errorf("%v not a int", v)
		}

		return i, nil
	default:
		return 0, fmt.Errorf("%v (%T) not a int", v, v)
	}
}


// SelectString is like Select but attempts to coerce the selection into a
// string. An error is returned if the coercion fails.
func (j Selecter) SelectString(sels ...interface{}) (string, error) {
	v, err := Select(j.V, sels...)
	if err != nil {
		return "", err
	}

	if v == nil {
		return "", fmt.Errorf("%w: %v", ErrNilValue, sels)
	}

	str, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%v not a string", v)
	}

	return str, nil
}

// SelectSlice is like Select but attempts to coerce the selection into a
// []interface{} (which gets converted into []Selecter). An error is
// returned if the coercion fails.
func (j Selecter) SelectSlice(sels ...interface{}) ([]Selecter, error) {
	v, err := Select(j.V, sels...)
	if err != nil {
		return nil, err
	}

	slcv, ok := v.([]interface{})
	if !ok {
		return nil, fmt.Errorf("%v not a slice", v)
	}

	slc := make([]Selecter, len(slcv))
	for i, v := range slcv {
		slc[i] = Selecter{V: v}
	}

	return slc, nil
}

// SelectMap is like Select but attempts to coerce the selection into a
// map[string]interface{} (which gets converted into map[string]Selecter).
// An error is returned if the coercion fails.
func (j Selecter) SelectMap(sels ...interface{}) (map[string]Selecter, error) {
	v, err := Select(j.V, sels...)
	if err != nil {
		return nil, err
	}

	mapv, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%v not a map", v)
	}

	mp := make(map[string]Selecter, len(mapv))
	for k, v := range mapv {
		mp[k] = Selecter{V: v}
	}

	return mp, nil
}

// SelectMapString is like Select but attempts to coerce the selection into a
// map[string]string (which gets converted into map[string]string). An error
// is returned if the coercion fails.
func (j Selecter) SelectMapString(sels ...interface{}) (map[string]string, error) {
	v, err := Select(j.V, sels...)
	if err != nil {
		return nil, err
	}

	mapv, ok := v.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%v not a slice", v)
	}

	mp := make(map[string]string, len(mapv))
	for k, v := range mapv {
		mp[k], err = Selecter{V: v}.SelectString()
		if err != nil {
			return nil, err
		}
	}

	return mp, nil
}

var ErrNilValue = errors.New("key exists but but nil value cannot be converted")

type ErrKeyNotPresent struct {
	Key interface{}
}

func (err ErrKeyNotPresent) Error() string {
	f := ""
	switch key := err.Key.(type) {
	case string:
		f = "key %q not found in object"
	case []int:
		switch len(key) {
		case 2: //
			f = "index %v out of bounds for array of len " +
				strconv.Itoa(key[1])

		case 3:
			f = "indeces %v out of bounds for array of len " +
				strconv.Itoa(key[2])

		default:
			f = "key %v (%T) not found in object"
		}
	default:
		// should not happen
		f = "key %v (%T) not found in object"
	}

	return fmt.Sprintf(f, err.Key)
}

func (err ErrKeyNotPresent) Is(arg error) bool {
	argv, ok := arg.(ErrKeyNotPresent)
	if !ok {
		argp, ok := arg.(*ErrKeyNotPresent)
		if !ok {
			return false
		}

		arg = *argp
	}

	// of arg is 0 value it's just a type check
	if argv == (ErrKeyNotPresent{}) {
		return true
	}

	return err == argv
}

// Select selects a value from a generic object created from passing
// interface{} into json.Unmarshal. sels have the following semantics:
//		string - select a value from a map[string]interface obj
//		[]string - filter a map[string]interface obj to only have the listed keys
//		int - select a value from a []interface{}
//		[]int if len 0 - noop
//		[]int if len 1 - select [n0:] from a []interface{}
//		[]int if len 2 - select [n0:n1] from a []interface{}
// All other combinations return an error
func Select(obj interface{}, sels ...interface{}) (interface{}, error) {

	if len(sels) == 0 {
		return obj, nil
	}

	var err error

	switch objv := obj.(type) {
	case map[string]interface{}:
		switch sel := sels[0].(type) {
		case string:
			v, ok := objv[sel]
			if !ok {
				return nil, ErrKeyNotPresent{sel}
			}

			return Select(v, sels[1:]...)

		case []string:
			ret := map[string]interface{}{}
			for _, seli := range sel {
				v, ok := objv[seli]
				if !ok {
					return nil, ErrKeyNotPresent{sel}
				}

				ret[seli], err = Select(v, sels[1:]...)
				if err != nil {
					return nil, err
				}
			}

			return ret, nil

		default:
			return nil, fmt.Errorf("cannot index object with %q", sels[0])
		}

	case []interface{}:

		switch sel := sels[0].(type) {
		case int:
			if sel < 0 || sel >= len(objv) {
				return nil, ErrKeyNotPresent{[]int{sel, len(objv)}}
			}

			return Select(objv[sel], sels[1:]...)

		case []int:
			start := 0
			end := len(objv)

			switch len(sel) {
			case 2:
				end = sel[1]
				fallthrough
			case 1:
				start = sel[0]
			case 0:
				// no op
			default:
				//len(sel) > 2
				return nil, fmt.Errorf("slice selector can have a max of 2 elements")
			}

			if start < 0 || start > len(objv) {
				return nil, ErrKeyNotPresent{append(sel, len(objv))}
			}

			if end < 0 || end > len(objv) {
				return nil, ErrKeyNotPresent{append(sel, len(objv))}
			}

			ret := make([]interface{}, end-start)
			for i, v := range objv[start:end] {
				ret[i], err = Select(v, sels[1:]...)
				if err != nil {
					return nil, err
				}
			}

			return ret, nil

		default:
			return nil, fmt.Errorf("cannot index array with %q", sels[0])
		}

	default:
		// the object we are selecting from is not a composite type
		return nil, fmt.Errorf("cannot select field %v of %v", sels[0], obj)
	}
}
