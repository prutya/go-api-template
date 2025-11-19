package repo

import (
	"context"
	"database/sql"
	"time"

	"github.com/uptrace/bun"

	"prutya/go-api-template/internal/models"
)

type UserRepo interface {
	Create(
		ctx context.Context,
		userId string,
		email string,
		passwordDigest string,
		emailVerificationExpiresAt time.Time,
		emailVerificationCooldownResetsAt time.Time,
	) error
	FindByID(ctx context.Context, userID string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailForUpdateNowait(ctx context.Context, email string) (*models.User, error)
	ResetPassword(ctx context.Context, userID string, newPasswordDigest string) error
	ChangePassword(ctx context.Context, userID string, newPasswordDigest string) error
	Delete(ctx context.Context, userID string) error
	StartEmailVerification(
		ctx context.Context,
		userId string,
		emailVerificationExpiresAt time.Time,
		emailVerificationCooldownResetsAt time.Time,
	) error
	IncrementEmailVerificationAttempts(ctx context.Context, userId string) error
	CompleteEmailVerification(ctx context.Context, userId string) error
	UpdateEmailVerificationOtpHmac(ctx context.Context, userId string, hmac []byte) error
	StartPasswordReset(
		ctx context.Context,
		userId string,
		passwordResetExpiresAt time.Time,
		passwordResetCooldownResetsAt time.Time,
	) error
	UpdatePasswordResetOtpHmac(ctx context.Context, userId string, hmac []byte) error
	IncrementPasswordResetAttempts(ctx context.Context, userId string) error
	StorePasswordResetTokenKey(
		ctx context.Context,
		userId string,
		passwordResetTokenPublicKey []byte,
	) error
}

type userRepo struct {
	db bun.IDB
}

func NewUserRepo(db bun.IDB) UserRepo {
	return &userRepo{
		db: db,
	}
}

func (r *userRepo) Create(
	ctx context.Context,
	userId string,
	email string,
	passwordDigest string,
	emailVerificationExpiresAt time.Time,
	emailVerificationCooldownResetsAt time.Time,
) error {
	user := &models.User{
		ID:                                userId,
		Email:                             email,
		PasswordDigest:                    passwordDigest,
		EmailVerificationExpiresAt:        sql.NullTime{Valid: true, Time: emailVerificationExpiresAt},
		EmailVerificationCooldownResetsAt: sql.NullTime{Valid: true, Time: emailVerificationCooldownResetsAt},
	}

	_, err := r.db.NewInsert().
		Model(user).
		Value("email_verification_last_requested_at", "now()").
		Exec(ctx)

	return err
}

func (r *userRepo) FindByID(ctx context.Context, userID string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().Model(user).Where("id = ?", userID).Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("lower(email) = lower(?)", email).
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) FindByEmailForUpdateNowait(ctx context.Context, email string) (*models.User, error) {
	user := new(models.User)
	err := r.db.NewSelect().
		Model(user).
		Where("lower(email) = lower(?)", email).
		For("UPDATE NOWAIT").
		Scan(ctx)

	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) ResetPassword(ctx context.Context, userID string, newPasswordDigest string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_digest = ?", newPasswordDigest).
		Set("password_reset_token_public_key = null").
		Set("updated_at = now()").
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

func (r *userRepo) ChangePassword(ctx context.Context, userID string, newPasswordDigest string) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_digest = ?", newPasswordDigest).
		Set("updated_at = now()").
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

func (r *userRepo) Delete(ctx context.Context, userID string) error {
	_, err := r.db.NewDelete().
		Model((*models.User)(nil)).
		Where("id = ?", userID).
		Exec(ctx)

	return err
}

func (r *userRepo) StartEmailVerification(
	ctx context.Context,
	userId string,
	emailVerificationExpiresAt time.Time,
	emailVerificationCooldownResetsAt time.Time,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("email_verification_otp_hmac = null").
		Set("email_verification_expires_at = ?", emailVerificationExpiresAt).
		Set("email_verification_otp_attempts = 0").
		Set("email_verification_cooldown_resets_at = ?", emailVerificationCooldownResetsAt).
		Set("email_verification_last_requested_at = now()").
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}

func (r *userRepo) IncrementEmailVerificationAttempts(
	ctx context.Context,
	userId string,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("email_verification_otp_attempts = email_verification_otp_attempts + 1").
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}

func (r *userRepo) CompleteEmailVerification(
	ctx context.Context,
	userId string,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("email_verified_at = now()").
		Set("email_verification_otp_hmac = null").
		Set("email_verification_expires_at = null").
		Set("email_verification_otp_attempts = 0").
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}

func (r *userRepo) UpdateEmailVerificationOtpHmac(
	ctx context.Context,
	userId string,
	hmac []byte,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("email_verification_otp_hmac = ?", hmac).
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}

func (r *userRepo) StartPasswordReset(
	ctx context.Context,
	userId string,
	passwordResetExpiresAt time.Time,
	passwordResetCooldownResetsAt time.Time,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_reset_otp_hmac = null").
		Set("password_reset_expires_at = ?", passwordResetExpiresAt).
		Set("password_reset_otp_attempts = 0").
		Set("password_reset_cooldown_resets_at = ?", passwordResetCooldownResetsAt).
		Set("password_reset_last_requested_at = now()").
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}

func (r *userRepo) UpdatePasswordResetOtpHmac(
	ctx context.Context,
	userId string,
	hmac []byte,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_reset_otp_hmac = ?", hmac).
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}

func (r *userRepo) IncrementPasswordResetAttempts(
	ctx context.Context,
	userId string,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_reset_otp_attempts = password_reset_otp_attempts + 1").
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}

func (r *userRepo) StorePasswordResetTokenKey(
	ctx context.Context,
	userId string,
	passwordResetTokenPublicKey []byte,
) error {
	_, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("password_reset_token_public_key = ?", passwordResetTokenPublicKey).
		Set("password_reset_otp_hmac = null").
		Set("password_reset_expires_at = null").
		Set("password_reset_otp_attempts = 0").
		Set("updated_at = now()").
		Where("id = ?", userId).
		Exec(ctx)

	return err
}
