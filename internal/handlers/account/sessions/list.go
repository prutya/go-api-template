package sessions

import (
	"net/http"
	"net/url"
	"strconv"
	"time"

	"prutya/go-api-template/internal/handlers/utils"
	"prutya/go-api-template/internal/services/authentication_service"
)

const defaultPageSize int = 50

type ListRequestQuery struct {
	PageSize *int    `query:"pageSize" validate:"omitempty,gte=1,lte=100"`
	Before   *string `query:"before" validate:"omitempty,uuid"`
}

type ListResponse struct {
	Items   []*ListResponseItem `json:"items"`
	HasMore bool                `json:"hasMore"`
}

type ListResponseItem struct {
	ID        string  `json:"id"`
	UserAgent *string `json:"userAgent"`
	IPAddress *string `json:"ipAddress"`
	ExpiresAt string  `json:"expiresAt"`
	CreatedAt string  `json:"createdAt"`
}

func NewSessionsListHandler(authenticationService authentication_service.AuthenticationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := &ListRequestQuery{}

		queryValues, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			utils.RenderError(w, r, utils.ErrInvalidQuery)
			return
		}

		// Get the page size from the query
		queryPageSize := queryValues.Get("pageSize")
		if queryPageSize != "" {
			pageSize, err := strconv.Atoi(queryPageSize)

			if err != nil {
				utils.RenderError(w, r, utils.ErrInvalidQuery)
				return
			}

			query.PageSize = &pageSize
		}

		// Get the start cursor from the query
		queryBefore := queryValues.Get("before")
		if queryBefore != "" {
			query.Before = &queryBefore
		}

		// Set the default page size
		if query.PageSize == nil {
			pageSize := defaultPageSize
			query.PageSize = &pageSize
		}

		// Validate the query params
		if err := utils.Validate.Struct(query); err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Get the sessions for the user
		sessions, hasMore, err := authenticationService.GetActiveSessionsForUser(
			r.Context(),
			utils.GetAccessTokenClaimsFromContext(r.Context()).UserID,
			*query.PageSize,
			query.Before,
		)
		if err != nil {
			utils.RenderError(w, r, err)
			return
		}

		// Create the response
		response := &ListResponse{
			Items:   make([]*ListResponseItem, len(sessions)),
			HasMore: hasMore,
		}

		for i, s := range sessions {
			var userAgent *string

			if s.UserAgent.Valid {
				userAgent = &s.UserAgent.String
			}

			var ipAddress *string

			if s.IPAddress.Valid {
				ipAddress = &s.IPAddress.String
			}

			response.Items[i] = &ListResponseItem{
				ID:        s.ID,
				UserAgent: userAgent,
				IPAddress: ipAddress,
				ExpiresAt: s.ExpiresAt.Format(time.RFC3339),
				CreatedAt: s.CreatedAt.Format(time.RFC3339),
			}
		}

		utils.RenderJson(w, r, response, http.StatusOK, nil)
	}
}
