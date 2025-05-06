package panic_check_test

import (
	"bytes"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"prutya/go-api-template/internal/handlers/panic_check"
	"prutya/go-api-template/internal/logger"
)

var _ = Describe("Panic_check", func() {
	It("panics", func() {
		handler := panic_check.NewHandler()
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/panic-check", bytes.NewBuffer([]byte("")))

		loggerInstance, err := logger.New("error", "json")
		if err != nil {
			panic(err)
		}

		r = r.WithContext(logger.NewContext(r.Context(), loggerInstance))

		Expect(func() {
			handler.ServeHTTP(w, r)
		}).To(Panic())
	})
})
