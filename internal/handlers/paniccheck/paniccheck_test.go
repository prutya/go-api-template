package paniccheck_test

import (
	"bytes"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"prutya/go-api-template/internal/handlers/paniccheck"
)

var _ = Describe("Paniccheck", func() {
	It("panics", func() {
		handler := paniccheck.NewHandler()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/paniccheck", bytes.NewBuffer([]byte("")))

		Expect(func() {
			handler.ServeHTTP(w, r)
		}).To(Panic())
	})
})
