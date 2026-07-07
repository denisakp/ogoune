package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/denisakp/ogoune/internal/domain"
	"github.com/denisakp/ogoune/internal/repository/fake"
	"github.com/denisakp/ogoune/internal/service"
	"github.com/pquerna/otp/totp"
)

func newTwoFactorTestServices(t *testing.T) (*service.AuthService, *service.TwoFactorService, *fake.UserRepository, *fake.TwoFactorResetTokenRepository) {
	t.Helper()
	userRepo := fake.NewUserRepository()
	jwtMgr := service.NewJWTManager("test-secret", "ogoune", time.Hour)
	authSvc := service.NewAuthService(userRepo, jwtMgr)
	tokenRepo := fake.NewTwoFactorResetTokenRepository()
	twoFactorSvc := service.NewTwoFactorService(authSvc, userRepo, tokenRepo, nil, "https://app.test")
	return authSvc, twoFactorSvc, userRepo, tokenRepo
}

func seedUser(t *testing.T, repo *fake.UserRepository, email string, twoFactor bool) *domain.User {
	t.Helper()
	u, err := repo.Create(context.Background(), &domain.User{
		Email:               email,
		HashedPassword:      "h",
		PasswordInitialized: true,
		TwoFactorEnabled:    twoFactor,
	})
	if err != nil {
		t.Fatalf("seed user: %v", err)
	}
	return u
}

func TestTwoFactor_Setup_StoresUnverifiedSecret(t *testing.T) {
	_, svc, userRepo, _ := newTwoFactorTestServices(t)
	u := seedUser(t, userRepo, "u@x.test", false)
	out, err := svc.Setup(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	if out.Secret == "" || out.OTPAuthURL == "" {
		t.Fatalf("setup must return secret + otpauth url, got %+v", out)
	}
	reloaded, _ := userRepo.FindByID(context.Background(), u.ID)
	if reloaded.TwoFactorSecret == "" || reloaded.TwoFactorEnabled {
		t.Fatalf("setup must persist unverified secret, got %+v", reloaded)
	}
}

func TestTwoFactor_Setup_AlreadyEnabledReturnsError(t *testing.T) {
	_, svc, userRepo, _ := newTwoFactorTestServices(t)
	u := seedUser(t, userRepo, "u@x.test", true)
	u.TwoFactorSecret = "JBSWY3DPEHPK3PXP"
	_ = userRepo.Update(context.Background(), u)
	_, err := svc.Setup(context.Background(), u.ID)
	if err != service.ErrTwoFactorAlreadyEnabled {
		t.Fatalf("expected ErrTwoFactorAlreadyEnabled, got %v", err)
	}
}

func TestTwoFactor_Verify_EnablesAndReturnsBackupCodes(t *testing.T) {
	_, svc, userRepo, _ := newTwoFactorTestServices(t)
	u := seedUser(t, userRepo, "u@x.test", false)
	out, err := svc.Setup(context.Background(), u.ID)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	code, err := totp.GenerateCode(out.Secret, time.Now())
	if err != nil {
		t.Fatalf("totp: %v", err)
	}
	codes, err := svc.Verify(context.Background(), u.ID, code)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if len(codes) != 10 {
		t.Fatalf("expected 10 backup codes, got %d", len(codes))
	}
	reloaded, _ := userRepo.FindByID(context.Background(), u.ID)
	if !reloaded.TwoFactorEnabled {
		t.Fatalf("user must be marked enabled after verify")
	}
}

func TestTwoFactor_Verify_BadCodeRejected(t *testing.T) {
	_, svc, userRepo, _ := newTwoFactorTestServices(t)
	u := seedUser(t, userRepo, "u@x.test", false)
	_, _ = svc.Setup(context.Background(), u.ID)
	_, err := svc.Verify(context.Background(), u.ID, "000000")
	if err != service.ErrTwoFactorBadCode {
		t.Fatalf("expected ErrTwoFactorBadCode, got %v", err)
	}
}

func TestTwoFactor_Disable_WipesSecret(t *testing.T) {
	_, svc, userRepo, _ := newTwoFactorTestServices(t)
	u := seedUser(t, userRepo, "u@x.test", false)
	out, _ := svc.Setup(context.Background(), u.ID)
	code, _ := totp.GenerateCode(out.Secret, time.Now())
	_, _ = svc.Verify(context.Background(), u.ID, code)

	code2, _ := totp.GenerateCode(out.Secret, time.Now())
	if err := svc.Disable(context.Background(), u.ID, code2); err != nil {
		t.Fatalf("disable: %v", err)
	}
	reloaded, _ := userRepo.FindByID(context.Background(), u.ID)
	if reloaded.TwoFactorEnabled || reloaded.TwoFactorSecret != "" {
		t.Fatalf("disable must wipe secret + flag, got %+v", reloaded)
	}
}

func TestTwoFactor_RequestReset_AntiEnumeration(t *testing.T) {
	_, svc, _, tokens := newTwoFactorTestServices(t)
	// Email not registered → still 202 (nil error), no token row.
	if err := svc.RequestReset(context.Background(), "ghost@x.test"); err != nil {
		t.Fatalf("anti-enum should not surface errors, got %v", err)
	}
	n, _ := tokens.CountRecentByUser(context.Background(), "ghost", time.Time{})
	if n != 0 {
		t.Fatalf("no token must be created for unknown email")
	}
}

func TestTwoFactor_RequestReset_HappyPathCreatesToken(t *testing.T) {
	_, svc, userRepo, tokens := newTwoFactorTestServices(t)
	u := seedUser(t, userRepo, "u@x.test", false)
	out, _ := svc.Setup(context.Background(), u.ID)
	code, _ := totp.GenerateCode(out.Secret, time.Now())
	_, _ = svc.Verify(context.Background(), u.ID, code)

	if err := svc.RequestReset(context.Background(), "u@x.test"); err != nil {
		t.Fatalf("request reset: %v", err)
	}
	n, _ := tokens.CountRecentByUser(context.Background(), u.ID, time.Now().Add(-time.Hour))
	if n != 1 {
		t.Fatalf("expected 1 token row, got %d", n)
	}
}

func TestTwoFactor_RequestReset_PerUserRateLimit(t *testing.T) {
	_, svc, userRepo, tokens := newTwoFactorTestServices(t)
	u := seedUser(t, userRepo, "u@x.test", false)
	out, _ := svc.Setup(context.Background(), u.ID)
	code, _ := totp.GenerateCode(out.Secret, time.Now())
	_, _ = svc.Verify(context.Background(), u.ID, code)

	for i := 0; i < 5; i++ {
		_ = svc.RequestReset(context.Background(), "u@x.test")
	}
	n, _ := tokens.CountRecentByUser(context.Background(), u.ID, time.Now().Add(-time.Hour))
	if n > 3 {
		t.Fatalf("rate limit broken: %d tokens", n)
	}
}

func TestTwoFactor_ConfirmReset_InvalidTokenReturns410(t *testing.T) {
	_, svc, _, _ := newTwoFactorTestServices(t)
	_, err := svc.ConfirmReset(context.Background(), "deadbeef")
	if err != service.ErrTwoFactorTokenInvalid {
		t.Fatalf("expected ErrTwoFactorTokenInvalid, got %v", err)
	}
}
