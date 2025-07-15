package repo

import "github.com/uptrace/bun"

type RepoFactory interface {
	NewAccessTokenRepo(db bun.IDB) AccessTokenRepo
	NewEmailSendAttemptRepo(db bun.IDB) EmailSendAttemptRepo
	NewEmailVerificationTokenRepo(db bun.IDB) EmailVerificationTokenRepo
	NewPasswordResetTokenRepo(db bun.IDB) PasswordResetTokenRepo
	NewRefreshTokenRepo(db bun.IDB) RefreshTokenRepo
	NewSessionRepo(db bun.IDB) SessionRepo
	NewUserRepo(db bun.IDB) UserRepo
}

type repoFactory struct{}

func NewRepoFactory() RepoFactory {
	return &repoFactory{}
}

func (f *repoFactory) NewAccessTokenRepo(db bun.IDB) AccessTokenRepo {
	return NewAccessTokenRepo(db)
}

func (f *repoFactory) NewEmailSendAttemptRepo(db bun.IDB) EmailSendAttemptRepo {
	return NewEmailSendAttemptRepo(db)
}

func (f *repoFactory) NewEmailVerificationTokenRepo(db bun.IDB) EmailVerificationTokenRepo {
	return NewEmailVerificationTokenRepo(db)
}

func (f *repoFactory) NewPasswordResetTokenRepo(db bun.IDB) PasswordResetTokenRepo {
	return NewPasswordResetTokenRepo(db)
}

func (f *repoFactory) NewRefreshTokenRepo(db bun.IDB) RefreshTokenRepo {
	return NewRefreshTokenRepo(db)
}

func (f *repoFactory) NewSessionRepo(db bun.IDB) SessionRepo {
	return NewSessionRepo(db)
}

func (f *repoFactory) NewUserRepo(db bun.IDB) UserRepo {
	return NewUserRepo(db)
}
