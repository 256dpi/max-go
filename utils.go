package max

import "strconv"

// ToInt will convert to int64.
func ToInt(atom Atom) int64 {
	switch v := atom.(type) {
	case int64:
		return v
	case float64:
		return int64(v)
	case string:
		n, _ := strconv.ParseInt(v, 10, 64)
		return n
	default:
		return 0
	}
}

// ToFloat will convert to float64.
func ToFloat(atom Atom) float64 {
	switch v := atom.(type) {
	case int64:
		return float64(v)
	case float64:
		return v
	case string:
		n, _ := strconv.ParseFloat(v, 64)
		return n
	default:
		return 0
	}
}

// ToString will convert to string.
func ToString(atom Atom) string {
	switch v := atom.(type) {
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case string:
		return v
	default:
		return ""
	}
}
