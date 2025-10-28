package db

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

// BoundingBox represents a bounding box with x, y, width, height
type BoundingBox struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// Value implements the driver.Valuer interface for database storage
func (bb BoundingBox) Value() (driver.Value, error) {
	return json.Marshal(bb)
}

// Scan implements the sql.Scanner interface for database retrieval
func (bb *BoundingBox) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into BoundingBox", value)
	}

	return json.Unmarshal(bytes, bb)
}
