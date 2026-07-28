// Package validate holds pure TI validation rules (no I/O, no mutation).
package validate

import (
	"strings"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

// ValidIOCType reports whether t is a known IOC type constant.
func ValidIOCType(t domain.IOCType) bool {
	switch t {
	case domain.IOCIP, domain.IOCDomain, domain.IOCURL, domain.IOCHash:
		return true
	default:
		return false
	}
}

// IOCShapeError describes why an IOC fails pre-normalization shape checks.
type IOCShapeError string

func (e IOCShapeError) Error() string { return string(e) }

const (
	ErrEmptyValue   IOCShapeError = "empty value"
	ErrUnknownType  IOCShapeError = "unknown ioc type"
	ErrEmptyActor   IOCShapeError = "empty actor name"
	ErrEmptyReport  IOCShapeError = "empty report link"
)

// CheckIOCShape validates type and non-empty value before normalization.
func CheckIOCShape(i domain.IOC) error {
	if !ValidIOCType(i.Type) {
		return ErrUnknownType
	}
	if strings.TrimSpace(i.Value) == "" {
		return ErrEmptyValue
	}
	return nil
}

// CheckActorName requires a non-empty actor label.
func CheckActorName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyActor
	}
	return nil
}

// CheckReportLink requires a non-empty report URL or identifier.
func CheckReportLink(link string) error {
	if strings.TrimSpace(link) == "" {
		return ErrEmptyReport
	}
	return nil
}
