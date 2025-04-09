package authentication_service

import "context"

func (s *authenticationService) Logout(ctx context.Context, accessTokenClaims *AccessTokenClaims) error {
	// Update the session directly with a subquery join in a single operation
	return s.sessionRepo.TerminateSessionByAccessTokenId(ctx, accessTokenClaims.ID)
}
