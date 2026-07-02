package firebase

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"time"

	firebaseAdmin "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"firebase.google.com/go/v4/messaging"
	"firebase.google.com/go/v4/storage"
	"google.golang.org/api/option"

	"github.com/isyll/go-grpc-starter/pkg/config"
	"github.com/isyll/go-grpc-starter/pkg/logger"
)

const fbStorageURLPrefix = "https://firebasestorage.googleapis.com/v0/b/%s/o/"

type Client struct {
	app           *firebaseAdmin.App
	auth          *auth.Client
	storage       *storage.Client
	config        *config.FirebaseConfig
	logger        *logger.Logger
	storageBucket string
}

func InitFirebase(
	env string,
	cfgs *config.Configs,
	logx *logger.Logger,
) (*Client, error) {
	fbConfig, err := config.LoadFirebaseConfig(env)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to load firebase config: %w",
			err,
		)
	}

	return NewClient(fbConfig, logx)
}

func NewClient(
	cfg *config.FirebaseConfig,
	logx *logger.Logger,
) (*Client, error) {
	ctx := context.Background()

	opt := option.WithAuthCredentialsFile(
		option.ServiceAccount,
		cfg.CredentialsFile,
	)

	fbConfig := &firebaseAdmin.Config{
		ProjectID:     cfg.ProjectID,
		StorageBucket: cfg.StorageBucket,
	}

	app, err := firebaseAdmin.NewApp(ctx, fbConfig, opt)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to initialize firebase app: %w",
			err,
		)
	}

	authClient, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to initialize firebase auth: %w",
			err,
		)
	}

	storageClient, err := app.Storage(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to initialize firebase storage: %w",
			err,
		)
	}

	return &Client{
		app:           app,
		auth:          authClient,
		storage:       storageClient,
		config:        cfg,
		logger:        logx,
		storageBucket: cfg.StorageBucket,
	}, nil
}

func (c *Client) CreateCustomToken(
	ctx context.Context,
	userID string,
) (string, error) {
	token, err := c.auth.CustomToken(ctx, userID)
	if err != nil {
		c.logger.Error(
			"Failed to create custom token",
			"user_id",
			userID,
			"error",
			err,
		)
		return "", fmt.Errorf("failed to create custom token: %w", err)
	}

	c.logger.Info("Custom token created", "user_id", userID)
	return token, nil
}

func (c *Client) VerifyToken(
	ctx context.Context,
	idToken string,
) (*auth.Token, error) {
	token, err := c.auth.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	return token, nil
}

func (c *Client) ValidateStorageURL(url string) bool {
	expectedPrefix := fmt.Sprintf(fbStorageURLPrefix, c.storageBucket)
	return strings.HasPrefix(url, expectedPrefix)
}

func (c *Client) IsAllowedExtension(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	return slices.Contains(c.config.AllowedExtensions, ext)
}

func (c *Client) GetMaxFileSizeBytes() int64 {
	return int64(c.config.MaxFileSizeMB * 1024 * 1024)
}

func (c *Client) GetMaxFileSizeMB() int {
	return c.config.MaxFileSizeMB
}

func (c *Client) GetAllowedExtensions() []string {
	return c.config.AllowedExtensions
}

func (c *Client) GetAvatarPath(accountID, filename string) string {
	return fmt.Sprintf(
		"%s/%s/%s",
		c.config.AvatarFolder,
		accountID,
		filename,
	)
}

func (c *Client) DeleteFile(ctx context.Context, path string) error {
	bucket, err := c.storage.DefaultBucket()
	if err != nil {
		return fmt.Errorf("failed to get bucket: %w", err)
	}

	obj := bucket.Object(path)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	c.logger.Info("File deleted from storage", "path", path)
	return nil
}

func (c *Client) FileExists(
	ctx context.Context,
	path string,
) (bool, error) {
	bucket, err := c.storage.DefaultBucket()
	if err != nil {
		return false, fmt.Errorf("failed to get bucket: %w", err)
	}

	obj := bucket.Object(path)
	_, err = obj.Attrs(ctx)
	if err != nil {
		if strings.Contains(err.Error(), "object doesn't exist") ||
			strings.Contains(err.Error(), "not found") {
			return false, nil
		}
		return false, fmt.Errorf(
			"failed to check file existence: %w",
			err,
		)
	}

	return true, nil
}

func (c *Client) ExtractPathFromURL(url string) (string, error) {
	prefix := fmt.Sprintf(fbStorageURLPrefix, c.storageBucket)
	if !strings.HasPrefix(url, prefix) {
		return "", fmt.Errorf("invalid storage URL")
	}

	path := strings.TrimPrefix(url, prefix)
	if idx := strings.Index(path, "?"); idx > 0 {
		path = path[:idx]
	}

	path = strings.ReplaceAll(path, "%2F", "/")
	return path, nil
}

func (c *Client) CreateCustomTokenWithClaims(
	ctx context.Context,
	userID string,
	claims map[string]any,
) (string, error) {
	token, err := c.auth.CustomTokenWithClaims(ctx, userID, claims)
	if err != nil {
		c.logger.Error(
			"Failed to create custom token with claims",
			"user_id",
			userID,
			"error",
			err,
		)
		return "", fmt.Errorf("failed to create custom token: %w", err)
	}

	return token, nil
}

func (c *Client) GenerateUploadToken(
	ctx context.Context,
	userID string,
) (*UploadToken, error) {
	expiresAt := time.Now().Add(c.config.UploadTokenExpiresIn)

	claims := map[string]any{
		"purpose":    "avatar_upload",
		"expires_at": expiresAt.Unix(),
	}

	token, err := c.CreateCustomTokenWithClaims(ctx, userID, claims)
	if err != nil {
		return nil, err
	}

	return &UploadToken{
		Token:     token,
		ExpiresAt: expiresAt,
		UserID:    userID,
	}, nil
}

type UploadToken struct {
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    string    `json:"user_id"`
}

func (c *Client) GetStorageBucket() string {
	return c.storageBucket
}

func (c *Client) GetProjectID() string {
	return c.config.ProjectID
}

func (c *Client) GetMessagingClient(
	ctx context.Context,
) (*messaging.Client, error) {
	client, err := c.app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to get messaging client: %w",
			err,
		)
	}
	return client, nil
}
