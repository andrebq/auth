package auth

import (
	"context"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

func CreateToken(ctx context.Context, db *sql.DB, login, token_type string, expiresAt time.Time) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	var uid string
	if err := lookupActiveLogin(ctx, &uid, db, login); err != nil {
		return "", err
	}
	genpass, err := randomSalt(20)
	if err != nil {
		return "", err
	}
	salt, salted, err := saltPassword(genpass)
	if err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx, `insert into db_tokens(token_id,
		token_type,
		uid,
		salt,
		token,
		created_at_unix,
		expires_at_unix) values (?, ?,?,?,?,?,?)`, id.String(), token_type, uid, salt, salted, time.Now().Unix(), expiresAt.Unix())
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v:%v", id.String(), base64.URLEncoding.EncodeToString(genpass)), nil
}

func TokenLogin(ctx context.Context, db *sql.DB, token string) (string, string, error) {
	tid, plain, err := splitToken(token)
	if err != nil {
		return "", "", err
	}
	var uid, tokenType string
	var salt, salted []byte
	now := time.Now().Unix()
	err = db.QueryRowContext(ctx,
		`select uid, token_type, salt, token from db_tokens where token_id = ? and created_at_unix <= ? and expires_at_unix > ?`, tid, now, now).Scan(
		&uid, &tokenType, &salt, &salted)
	if err != nil {
		return "", "", err
	}
	if !validatePasswd(salt, salted, plain) {
		return "", "", errors.New("auth: credentials not found or invalid")
	}
	return uid, tokenType, nil
}

func RevokeToken(ctx context.Context, db *sql.DB, tokenID string) error {
	changes, err := db.ExecContext(ctx, `update db_tokens set expires_at_unix = created_at_unix where token_id = ?`, tokenID)
	if err != nil {
		return err
	}
	rows, err := changes.RowsAffected()
	if err != nil {
		return err
	}
	if rows != 1 {
		return errors.New("auth: invalid token format")
	}
	return nil
}

func ExtractTokenID(t string) (string, error) {
	i := strings.Index(t, ":")
	switch {
	case i < 0:
		return t, nil
	case i == 0:
		return "", errors.New("auth: invalid token format")
	default:
		return t[:i], nil
	}
}

func splitToken(t string) (string, []byte, error) {
	parts := strings.SplitN(t, ":", 2)
	if len(parts) != 2 {
		return "", nil, errors.New("auth: invalid token format")
	}
	plain, err := base64.URLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", nil, errors.New("auth: invalid token format")
	}
	return parts[0], plain, nil
}
