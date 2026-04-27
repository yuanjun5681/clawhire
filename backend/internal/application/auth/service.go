package auth

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"

	"github.com/yuanjun5681/clawhire/backend/internal/domain/account"
	infraauth "github.com/yuanjun5681/clawhire/backend/internal/infrastructure/auth"
	"github.com/yuanjun5681/clawhire/backend/internal/shared/apierr"
)

// accountIDPattern 限定 human 注册时允许的字面量：英文数字加 _ - ., 3-64 位。
// 避免与 agent 命名空间冲突：human 账号统一加 "acct_human_" 前缀。
var accountIDPattern = regexp.MustCompile(`^[a-zA-Z0-9_\-.]{3,64}$`)

const humanAccountPrefix = "acct_human_"

// Service 聚合 human 账号注册、登录、会话签发。
type Service struct {
	accounts    account.Repository
	issuer      *infraauth.JWTIssuer
	bcryptCost  int
	minPassword int
	now         func() time.Time
}

type Options struct {
	Accounts    account.Repository
	Issuer      *infraauth.JWTIssuer
	BcryptCost  int
	MinPassword int
	Now         func() time.Time
}

func NewService(opt Options) *Service {
	cost := opt.BcryptCost
	if cost <= 0 {
		cost = bcrypt.DefaultCost
	}
	minPwd := opt.MinPassword
	if minPwd <= 0 {
		minPwd = 8
	}
	now := opt.Now
	if now == nil {
		now = time.Now
	}
	return &Service{
		accounts:    opt.Accounts,
		issuer:      opt.Issuer,
		bcryptCost:  cost,
		minPassword: minPwd,
		now:         now,
	}
}

// RegisterHumanInput 是注册请求的语义载荷。
type RegisterHumanInput struct {
	AccountID   string
	DisplayName string
	Password    string
}

type AuthResult struct {
	Account   *account.Account
	Token     string
	ExpiresAt time.Time
}

// Register 创建一个 human 账号并返回会话 token。
func (s *Service) Register(ctx context.Context, in RegisterHumanInput) (*AuthResult, error) {
	accountID := strings.TrimSpace(in.AccountID)
	if !accountIDPattern.MatchString(accountID) {
		return nil, apierr.New(apierr.CodeInvalidRequest, "account id must be 3-64 chars, letters/digits/_-. only")
	}
	accountID = humanAccountPrefix + accountID

	displayName := strings.TrimSpace(in.DisplayName)
	if displayName == "" {
		return nil, apierr.New(apierr.CodeInvalidRequest, "display name is required")
	}
	if len(in.Password) < s.minPassword {
		return nil, apierr.New(apierr.CodeInvalidRequest, "password too short")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(in.Password), s.bcryptCost)
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternalError, "hash password", err)
	}

	now := s.now().UTC()
	acc := &account.Account{
		AccountID:    accountID,
		Type:         account.TypeHuman,
		DisplayName:  displayName,
		Status:       account.StatusActive,
		PasswordHash: string(hash),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := s.accounts.Insert(ctx, acc); err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, apierr.New(apierr.CodeConflict, "account id already taken")
		}
		return nil, apierr.Wrap(apierr.CodeInternalError, "insert account", err)
	}

	return s.issue(acc)
}

// Login 按 accountId + 密码 发放会话 token。
func (s *Service) Login(ctx context.Context, accountID, password string) (*AuthResult, error) {
	accountID = strings.TrimSpace(accountID)
	if !strings.HasPrefix(accountID, humanAccountPrefix) {
		accountID = humanAccountPrefix + accountID
	}
	if accountID == humanAccountPrefix || password == "" {
		return nil, apierr.New(apierr.CodeUnauthorized, "invalid credentials")
	}
	acc, err := s.accounts.FindByID(ctx, accountID)
	if err != nil {
		if errors.Is(err, account.ErrAccountNotFound) {
			return nil, apierr.New(apierr.CodeUnauthorized, "invalid credentials")
		}
		return nil, apierr.Wrap(apierr.CodeInternalError, "find account", err)
	}
	if acc.Type != account.TypeHuman {
		return nil, apierr.New(apierr.CodeForbidden, "only human accounts can sign in")
	}
	if acc.Status != account.StatusActive {
		return nil, apierr.New(apierr.CodeForbidden, "account is not active")
	}
	if acc.PasswordHash == "" {
		return nil, apierr.New(apierr.CodeUnauthorized, "invalid credentials")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(acc.PasswordHash), []byte(password)); err != nil {
		return nil, apierr.New(apierr.CodeUnauthorized, "invalid credentials")
	}
	return s.issue(acc)
}

func (s *Service) issue(acc *account.Account) (*AuthResult, error) {
	token, exp, err := s.issuer.Issue(acc.AccountID, string(acc.Type))
	if err != nil {
		return nil, apierr.Wrap(apierr.CodeInternalError, "issue token", err)
	}
	return &AuthResult{Account: acc, Token: token, ExpiresAt: exp}, nil
}
