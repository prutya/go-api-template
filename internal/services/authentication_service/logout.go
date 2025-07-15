package authentication_service

import "context"

// NOTE: I am not using transactions here, because it's just a single write
// operation
func (s *authenticationService) Logout(ctx context.Context, accessTokenClaims *AccessTokenClaims) error {
	sessionRepo := s.repoFactory.NewSessionRepo(s.db)

	// Update the session directly with a subquery join in a single operation
	return sessionRepo.TerminateSessionByAccessTokenId(ctx, accessTokenClaims.ID)
}
