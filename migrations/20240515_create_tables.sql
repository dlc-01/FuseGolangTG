-- +goose Up
-- SQL in section 'Up' is executed when this migration is applied

CREATE TABLE IF NOT EXISTS inodes (
                                      id SERIAL PRIMARY KEY,
                                      uid INT NOT NULL,
                                      gid INT NOT NULL,
                                      mode INT NOT NULL,
                                      mtime_ns BIGINT NOT NULL,
                                      atime_ns BIGINT NOT NULL,
                                      ctime_ns BIGINT NOT NULL,
                                      size BIGINT NOT NULL DEFAULT 0,
                                      rdev BIGINT NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS telegram_messages (
                                                 id SERIAL PRIMARY KEY,
                                                 inode INT NOT NULL,
                                                 message_id VARCHAR(255) NOT NULL,
    FOREIGN KEY (inode) REFERENCES inodes (id) ON DELETE CASCADE
    );

CREATE TABLE IF NOT EXISTS contents (
                                        rowid SERIAL PRIMARY KEY,
                                        name TEXT NOT NULL,
                                        inode INT NOT NULL,
                                        parent_inode INT NOT NULL,
                                        UNIQUE (name, parent_inode),
    FOREIGN KEY (inode) REFERENCES inodes (id) ON DELETE CASCADE
    );

-- +goose Down
-- SQL section 'Down' is executed when this migration is rolled back

DROP TABLE IF EXISTS contents;
DROP TABLE IF EXISTS telegram_messages;
DROP TABLE IF EXISTS inodes;
