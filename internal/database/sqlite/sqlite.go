package sqlite

import (
	"database/sql"
	"errors"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"geodb-example/internal/utils"
)

type SqliteDB struct {
	*sql.DB
	version string
	mu      sync.RWMutex
}

// Open подключается к существующей БД
func Open(dsn string) (*SqliteDB, error) {
	s := strings.TrimPrefix(dsn, "file://") // Удаляем схему
	path, params, _ := strings.Cut(s, "?")  // Получаем путь до БД и параметры

	// Проверяем: существует ли БД?
	fullPath, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	if !utils.IsFileExists(fullPath) {
		return nil, errors.New("database doesn't exist")
	}

	// Парсируем параметры и накидываем дополнительные
	queries, err := url.ParseQuery(params)
	if err != nil {
		return nil, err
	}
	queries.Set("_fk", "1")

	// Открываем БД
	db, err := sql.Open("sqlite3", fullPath+"?"+queries.Encode())
	if err != nil {
		return nil, err
	}

	d := &SqliteDB{DB: db}
	return d, nil
}
