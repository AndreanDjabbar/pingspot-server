package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)


func StrPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func BoolPtrOrNil(b bool) *bool {
	if !b {
		return nil
	}
	return &b
}

func IntPtrOrNil(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

func Int64PtrOrNil(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func UintToStrPtr(u uint) *string {
	if u == 0 {
		return nil
	}
	s := strconv.FormatUint(uint64(u), 10)
	return &s
}

func StringToInt(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	value, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to int: %w", err)
	}
	return value, nil
}

func StringToUint(s string) (uint, error) {
	if s == "" {
		return 0, nil
	}
	value, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to convert string to uint: %w", err)
	}
	return uint(value), nil
}

func StringToTimePtr(s string) (*time.Time, error) {
	if s == "" {
		return nil, nil
	}
	layout := time.RFC3339
	value, err := time.Parse(layout, s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to time.Time: %w", err)
	}
	return &value, nil
}

func StringToBool(s string) (bool, error) {
	if s == "" {
		return false, nil
	}
	value, err := strconv.ParseBool(s)
	if err != nil {
		return false, fmt.Errorf("failed to convert string to bool: %w", err)
	}
	return value, nil
}

func StringToFloat64(s string) (float64, error) {
	if s == "" {
		return 0.0, nil
	}
	value, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0, fmt.Errorf("failed to convert string to float64: %w", err)
	}

	return value, nil
}

func StringPtrToObjectIDPtr(s *string) (*primitive.ObjectID, error) {
	if s == nil || *s == "" {
		return nil, nil
	}
	objectID, err := primitive.ObjectIDFromHex(*s)
	if err != nil {
		return nil, fmt.Errorf("failed to convert string to ObjectID: %w", err)
	}
	return &objectID, nil
}

func ObjectIDPtrToStringPtr(id *primitive.ObjectID) *string {
	if id == nil {
		return nil
	}
	s := id.Hex()
	return &s
}

func MustJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}
