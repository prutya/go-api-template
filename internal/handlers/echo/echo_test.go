package echo_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"prutya/go-api-template/internal/handlers/echo"
	"prutya/go-api-template/internal/logger"
)

var _ = Describe("Echo", func() {
	var handler http.HandlerFunc
	var requestBody []byte
	var w *httptest.ResponseRecorder
	var r *http.Request

	JustBeforeEach(func() {

		handler = echo.NewHandler()
		w = httptest.NewRecorder()
		r = httptest.NewRequest("GET", "/echo", bytes.NewBuffer(requestBody))

		loggerInstance, err := logger.New("fatal", "2006-01-02T15:04:05Z07:00")
		if err != nil {
			panic(err)
		}

		r = r.WithContext(logger.NewContext(r.Context(), loggerInstance))

		handler.ServeHTTP(w, r)
	})

	Context("when the request body is not a valid JSON", func() {
		BeforeEach(func() {
			requestBody = []byte(`yolo`)
		})

		It("returns a 400", func() {
			Expect(w.Code).To(Equal(http.StatusBadRequest))
		})

		It("returns an invalid_json error", func() {
			Expect(w.Body.Bytes()).To(MatchJSON(`
				{
					"error": "invalid_json"
				}
			`))
		})
	})

	Context("when the request body is a valid JSON", func() {
		Context("when the message is missing", func() {
			BeforeEach(func() {
				requestBody = []byte(`{}`)
			})

			It("returns a 422", func() {
				Expect(w.Code).To(Equal(http.StatusUnprocessableEntity))
			})

			It("returns an invalid_params error", func() {
				Expect(w.Body.Bytes()).To(MatchJSON(`
					{
						"error": "invalid_params",
						"details": [
							{
								"subject": "message",
								"constraint": "required",
								"args": ""
							}
						]
					}
				`))
			})
		})

		Context("when the message is not a string", func() {
			BeforeEach(func() {
				requestBody = []byte(`{"message": 123}`)
			})

			It("returns a 400 with an invalid_json error", func() {
				Expect(w.Code).To(Equal(http.StatusBadRequest))
			})

			It("returns an invalid_json error", func() {
				Expect(w.Body.Bytes()).To(MatchJSON(`
					{
						"error": "invalid_json"
					}
				`))
			})
		})

		Context("when the message is a string", func() {
			Context("when the message is too short", func() {
				BeforeEach(func() {
					requestBody = []byte(`{"message": "a"}`)
				})

				It("returns a 422", func() {
					Expect(w.Code).To(Equal(http.StatusUnprocessableEntity))
				})

				It("returns an invalid_params error", func() {
					Expect(w.Body.Bytes()).To(MatchJSON(`
						{
							"error": "invalid_params",
							"details": [
								{
									"subject": "message",
									"constraint": "gte",
									"args": "2"
								}
							]
						}
					`))
				})
			})

			Context("when the message is too long", func() {
				BeforeEach(func() {
					requestBody = []byte(`{"message": "12345678901234567"}`)
				})

				It("returns a 422", func() {
					Expect(w.Code).To(Equal(http.StatusUnprocessableEntity))
				})

				It("returns an invalid_params error", func() {
					Expect(w.Body.Bytes()).To(MatchJSON(`
						{
							"error": "invalid_params",
							"details": [
								{
									"subject": "message",
									"constraint": "lte",
									"args": "16"
								}
							]
						}
					`))
				})
			})
		})
	})
})
