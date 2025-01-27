package panic_check_test

import (
	"bytes"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"prutya/go-api-template/internal/handlers/panic_check"
)

var _ = Describe("Panic_check", func() {
	It("panics", func() {
		handler := panic_check.NewHandler()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/panic-check", bytes.NewBuffer([]byte("")))

		Expect(func() {
			handler.ServeHTTP(w, r)
		}).To(Panic())
	})
})
