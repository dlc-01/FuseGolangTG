package postgres

import (
	"database/sql"
	"fmt"
	"github.com/dlc-01/config"
	"github.com/dlc-01/domain"
	"github.com/dlc-01/migrations"
	"github.com/dlc-01/ports"
	"io/ioutil"
	"time"

	_ "github.com/jackc/pgx/v4/stdlib"
)

type PostgresAdapter struct {
	db *sql.DB
}

func NewPostgresAdapter(cfg *config.Config) (ports.FileStoragePort, error) {
	db, err := sql.Open("pgx", cfg.PostgresURL)
	if err != nil {
		return nil, err
	}

	adapter := &PostgresAdapter{db: db}

	migrations.RunMigrations(cfg.PostgresURL, cfg.MigrationDir)

	return adapter, nil
}

func (s *PostgresAdapter) Lookup(parentInode uint64, name string) (domain.File, error) {
	row := s.db.QueryRow(`
		SELECT i.id, i.uid, i.gid, i.mode, i.mtime_ns, i.atime_ns, i.ctime_ns, i.size, i.rdev, tm.message_id
		FROM contents c
		JOIN inodes i ON c.inode = i.id
		LEFT JOIN telegram_messages tm ON i.id = tm.inode
		WHERE c.parent_inode = $1 AND c.name = $2`, parentInode, name)

	var file domain.File
	var mtimeNs, atimeNs, ctimeNs int64
	err := row.Scan(&file.ID, &file.Uid, &file.Gid, &file.Mode, &mtimeNs, &atimeNs, &ctimeNs, &file.Size, &file.Rdev, &file.TelegramID)
	if err != nil {
		return file, err
	}

	file.Mtime = time.Unix(0, mtimeNs)
	file.Atime = time.Unix(0, atimeNs)
	file.Ctime = time.Unix(0, ctimeNs)

	return file, nil
}

func (s *PostgresAdapter) ReadDirAll(parentInode uint64) ([]domain.File, error) {
	rows, err := s.db.Query(`
		SELECT i.id, i.uid, i.gid, i.mode, i.mtime_ns, i.atime_ns, i.ctime_ns, i.size, i.rdev, tm.message_id
		FROM contents c
		JOIN inodes i ON c.inode = i.id
		LEFT JOIN telegram_messages tm ON i.id = tm.inode
		WHERE c.parent_inode = $1`, parentInode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []domain.File
	for rows.Next() {
		var file domain.File
		var mtimeNs, atimeNs, ctimeNs int64
		err := rows.Scan(&file.ID, &file.Uid, &file.Gid, &file.Mode, &mtimeNs, &atimeNs, &ctimeNs, &file.Size, &file.Rdev, &file.TelegramID)
		if err != nil {
			return nil, err
		}
		file.Mtime = time.Unix(0, mtimeNs)
		file.Atime = time.Unix(0, atimeNs)
		file.Ctime = time.Unix(0, ctimeNs)
		files = append(files, file)
	}

	return files, nil
}

func (s *PostgresAdapter) Create(parentInode uint64, name string, mode uint32, uid uint32, gid uint32) (domain.File, error) {
	now := time.Now().UnixNano()

	var id uint64
	err := s.db.QueryRow(`
		INSERT INTO inodes (uid, gid, mode, mtime_ns, atime_ns, ctime_ns, size, rdev)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`, uid, gid, mode, now, now, now, 0, 0).Scan(&id)
	if err != nil {
		return domain.File{}, err
	}

	_, err = s.db.Exec(`
		INSERT INTO contents (name, inode, parent_inode)
		VALUES ($1, $2, $3)`, name, id, parentInode)
	if err != nil {
		return domain.File{}, err
	}

	return domain.File{
		ID:    id,
		Name:  name,
		Uid:   uid,
		Gid:   gid,
		Mode:  mode,
		Size:  0,
		Mtime: time.Unix(0, now),
		Atime: time.Unix(0, now),
		Ctime: time.Unix(0, now),
	}, nil
}

func (s *PostgresAdapter) Remove(parentInode uint64, name string) error {
	var id uint64
	err := s.db.QueryRow(`
		SELECT inode FROM contents
		WHERE parentInode = $1 AND name = $2`, parentInode, name).Scan(&id)
	if err != nil {
		return err
	}

	_, err = s.db.Exec(`
		DELETE FROM inodes WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostgresAdapter) UpdateTelegramID(inode uint64, telegramID string) error {
	_, err := s.db.Exec(`
		INSERT INTO telegram_messages (inode, message_id)
		VALUES ($1, $2)
		ON CONFLICT (inode) DO UPDATE SET message_id = $2`, inode, telegramID)
	return err
}

func (s *PostgresAdapter) SaveMapping(fileID string, messageID int) error {
	_, err := s.db.Exec(`
		INSERT INTO telegram_file_mapping (file_id, message_id)
		VALUES ($1, $2)
		ON CONFLICT (file_id) DO UPDATE SET message_id = $2`, fileID, messageID)
	return err
}

func (s *PostgresAdapter) FindMessageIDByFileID(fileID string) (int, error) {
	var messageID int
	err := s.db.QueryRow(`
		SELECT message_id FROM telegram_file_mapping
		WHERE file_id = $1`, fileID).Scan(&messageID)
	if err != nil {
		return 0, err
	}
	return messageID, nil
}

func (s *PostgresAdapter) Read(inode uint64, offset int64, size int) ([]byte, error) {
	var fileID string
	err := s.db.QueryRow(`SELECT message_id FROM telegram_messages WHERE inode = $1`, inode).Scan(&fileID)
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(fileID)
	if err != nil {
		return nil, err
	}
	if int(offset) > len(data) {
		return nil, fmt.Errorf("offset out of range")
	}
	end := offset + int64(size)
	if end > int64(len(data)) {
		end = int64(len(data))
	}
	return data[offset:end], nil
}

func (s *PostgresAdapter) Write(inode uint64, offset int64, data []byte) error {
	var fileID string
	err := s.db.QueryRow(`SELECT message_id FROM telegram_messages WHERE inode = $1`, inode).Scan(&fileID)
	if err != nil {
		return err
	}
	fileData, err := ioutil.ReadFile(fileID)
	if err != nil {
		return err
	}
	newData := append(fileData[:offset], data...)
	return ioutil.WriteFile(fileID, newData, 0644)
}
