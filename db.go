package auth

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"runtime"

	"github.com/google/uuid"
	"golang.org/x/crypto/argon2"

	"github.com/uptrace/bun/driver/sqliteshim"
)

func OpenDir(ctx context.Context, dir string) (*sql.DB, error) {
	db, err := sql.Open(sqliteshim.ShimName, fmt.Sprintf("file:%v?_pragma=foreign_keys(1)", filepath.Join(dir, "users.db")))
	if err != nil {
		return nil, err
	}
	err = InitDB(ctx, db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func OpenMemory(ctx context.Context) (*sql.DB, error) {
	db, err := sql.Open(sqliteshim.ShimName, "file::memory:?cache=shared")
	if err != nil {
		return nil, err
	}
	err = InitDB(ctx, db)
	if err != nil {
		return nil, err
	}
	return db, nil
}

// InitDB creates the required tables on the given db object
func InitDB(ctx context.Context, db *sql.DB) error {
	return execCmds(ctx, db, []string{
		`create table if not exists db_users(
			uid text not null,
			login text not null,
			salt blob not null,
			passwd blob not null,
			active integer not null,
			primary key(uid),
			unique(login))`,
		`create table if not exists db_tokens(token_id text not null,
			token_type text not null,
			uid text not null,
			salt blob not null,
			token blob not null,
			created_at_unix integer not null,
			expires_at_unix integer not null,
			primary key(token_id))`,
	})
}

// RegisterUser with the given login and password
func RegisterUser(ctx context.Context, db *sql.DB, login string, passwd []byte) (string, error) {
	uid, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	salt, salted, err := saltPassword(passwd)
	if err != nil {
		return "", err
	}
	_, err = db.ExecContext(ctx, `insert into db_users(uid, login, salt, passwd, active) values (?, ?, ?, ?, ?)`,
		uid.String(), login, salt, salted, 1)
	if err != nil {
		return "", err
	}
	return uid.String(), nil
}

// ReplacePassword of given user
func ReplacePassword(ctx context.Context, db *sql.DB, login string, newpass []byte) error {
	var uid string
	err := db.QueryRowContext(ctx, `select uid from db_users where login = ?`, login).Scan(&uid)
	if err != nil {
		return err
	}
	salt, salted, err := saltPassword(newpass)
	if err != nil {
		return err
	}
	_, err = db.ExecContext(ctx, `update db_users set passwd = ?, salt = ? where uid = ?`, salted, salt, uid)
	return err
}

// Login user
func Login(ctx context.Context, db *sql.DB, login string, plainPass []byte) (string, error) {
	var uid string
	var salted []byte
	var salt []byte
	err := db.QueryRowContext(ctx, `select uid, salt, passwd from db_users where login = ? and active = 1`, login).Scan(&uid, &salt, &salted)
	if err != nil {
		return "", err
	}
	if !validatePasswd(salt, salted, plainPass) {
		return "", errors.New("auth: credentials not found or invalid")
	}
	return uid, nil
}

func lookupActiveLogin(ctx context.Context, uid *string, db *sql.DB, login string) error {
	return db.QueryRowContext(ctx, `select uid from db_users where login = ? and active = 1`, login).Scan(uid)
}

func randomSalt(sz int) ([]byte, error) {
	buf := make([]byte, sz)
	_, err := io.ReadFull(rand.Reader, buf[:])
	if err != nil {
		return nil, err
	}
	return buf[:], nil
}

func saltPassword(plain []byte) ([]byte, []byte, error) {
	salt, err := randomSalt(8)
	if err != nil {
		return nil, nil, err
	}
	return salt, argon2.IDKey(plain, salt, 2, 32*1024, uint8(runtime.NumCPU()), 16), nil
}

func validatePasswd(salt, salted, plain []byte) bool {
	key := argon2.IDKey(plain, salt, 2, 32*1024, uint8(runtime.NumCPU()), 16)
	return subtle.ConstantTimeCompare(key, salted) == 1
}

func execCmds(ctx context.Context, db *sql.DB, cmds []string) error {
	for _, c := range cmds {
		_, err := db.ExecContext(ctx, c)
		if err != nil {
			return err
		}
	}
	return nil
}
