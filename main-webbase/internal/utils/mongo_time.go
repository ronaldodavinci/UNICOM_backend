package utils

import (
    "time"
    "go.mongodb.org/mongo-driver/v2/bson"
)

func ExtractTime(m bson.M, key string) (time.Time, bool) {
	v, ok := m[key]
	if !ok {
		return time.Time{}, false
	}
	switch tv := v.(type) {
	case time.Time:
		return tv, true
	case bson.DateTime:
		return tv.Time(), true
	case string:
		if t, err := time.Parse(time.RFC3339Nano, tv); err == nil {
			return t, true
		}
		if t, err := time.Parse(time.RFC3339, tv); err == nil {
			return t, true
		}
	}
	return time.Time{}, false
}
