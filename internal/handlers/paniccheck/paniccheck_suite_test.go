package paniccheck_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPaniccheck(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Paniccheck Suite")
}