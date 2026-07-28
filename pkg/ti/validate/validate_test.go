package validate

import (
	"testing"

	"github.com/butbeautifulv/veil/pkg/ti/domain"
)

func TestValidIOCType(t *testing.T) {
	for _, typ := range []domain.IOCType{domain.IOCIP, domain.IOCDomain, domain.IOCURL, domain.IOCHash} {
		if !ValidIOCType(typ) {
			t.Fatalf("%q should be valid", typ)
		}
	}
	if ValidIOCType(domain.IOCType("email")) {
		t.Fatal("unknown type")
	}
}

func TestCheckIOCShape(t *testing.T) {
	if err := CheckIOCShape(domain.IOC{Type: domain.IOCIP, Value: "1.2.3.4"}); err != nil {
		t.Fatal(err)
	}
	if err := CheckIOCShape(domain.IOC{Type: domain.IOCIP, Value: "  "}); err != ErrEmptyValue {
		t.Fatalf("got %v", err)
	}
	if err := CheckIOCShape(domain.IOC{Type: "x", Value: "a"}); err != ErrUnknownType {
		t.Fatalf("got %v", err)
	}
}

func TestIOCShapeError_Error(t *testing.T) {
	if ErrEmptyValue.Error() != "empty value" {
		t.Fatalf("got %q", ErrEmptyValue.Error())
	}
	if ErrUnknownType.Error() != "unknown ioc type" {
		t.Fatalf("got %q", ErrUnknownType.Error())
	}
	if ErrEmptyActor.Error() != "empty actor name" {
		t.Fatalf("got %q", ErrEmptyActor.Error())
	}
	if ErrEmptyReport.Error() != "empty report link" {
		t.Fatalf("got %q", ErrEmptyReport.Error())
	}
}

func TestCheckActorName_and_ReportLink(t *testing.T) {
	if CheckActorName("APT29") != nil {
		t.Fatal("actor")
	}
	if CheckActorName(" ") != ErrEmptyActor {
		t.Fatal("empty actor")
	}
	if CheckReportLink("https://x") != nil {
		t.Fatal("report")
	}
	if CheckReportLink("") != ErrEmptyReport {
		t.Fatal("empty report")
	}
}
