package secureform

import (
	"mime/multipart"
	"strconv"
)

// File represents a form input element such as <input type="file" />
type File struct {
	multipart.File
	*multipart.FileHeader
}

// Type represents an arbitrary interface for parsing
type Type interface {
	Set(value string) error
}

func validateInt(value int64, tag *parserTag) error {
	if tag.Min != "" {
		min, err := strconv.ParseInt(tag.Min, 0, 64)
		if err != nil {
			return err
		}
		if value < min {
			return ErrValidMin
		}
	}
	if tag.Max != "" {
		max, err := strconv.ParseInt(tag.Max, 0, 64)
		if err != nil {
			return err
		}
		if value > max {
			return ErrValidMax
		}
	}
	return nil
}

func validateUint(value uint64, tag *parserTag) error {
	if tag.Min != "" {
		min, err := strconv.ParseUint(tag.Min, 0, 64)
		if err != nil {
			return err
		}
		if value < min {
			return ErrValidMin
		}
	}
	if tag.Max != "" {
		max, err := strconv.ParseUint(tag.Max, 0, 64)
		if err != nil {
			return err
		}
		if value > max {
			return ErrValidMax
		}
	}
	return nil
}

func validateFloat(value float64, tag *parserTag) error {
	if tag.Min != "" {
		min, err := strconv.ParseFloat(tag.Min, 64)
		if err != nil {
			return err
		}
		if value < min {
			return ErrValidMin
		}
	}
	if tag.Max != "" {
		max, err := strconv.ParseFloat(tag.Max, 64)
		if err != nil {
			return err
		}
		if value > max {
			return ErrValidMax
		}
	}
	return nil
}

func validateString(value string, tag *parserTag, rootmax int) error {
	if tag.Min != "" {
		min, err := strconv.ParseInt(tag.Min, 0, 64)
		if err != nil {
			return err
		}
		if len(value) < int(min) {
			return ErrValidMin
		}
	}
	if tag.Max != "" {
		max, err := strconv.ParseInt(tag.Max, 0, 64)
		if err != nil {
			return err
		}
		if len(value) > int(max) {
			return ErrValidMax
		}
	}
	if len(value) > rootmax {
		return ErrValidMax
	}
	return nil
}
