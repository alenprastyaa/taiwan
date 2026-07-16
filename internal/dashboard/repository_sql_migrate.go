package dashboard

import (
	"context"
	"fmt"
)

func (r *SQLRepository) migrate(ctx context.Context) error {
	var statements []string
	switch r.dialect {
	case dialectPostgres:
		statements = postgresSchema()
	case dialectMySQL:
		statements = mysqlSchema()
	default:
		return fmt.Errorf("unsupported sql dialect %q", r.dialect)
	}

	for _, statement := range statements {
		if _, err := r.db.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}
	return nil
}

func postgresSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(64) PRIMARY KEY,
			username VARCHAR(191) NOT NULL UNIQUE,
			name VARCHAR(191) NOT NULL,
			email VARCHAR(191) NOT NULL DEFAULT '',
			phone VARCHAR(64) NOT NULL DEFAULT '',
			role VARCHAR(32) NOT NULL,
			password_hash BYTEA NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS clients (
			id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64),
			name VARCHAR(191) NOT NULL,
			email VARCHAR(191) NOT NULL DEFAULT '',
			phone VARCHAR(64) NOT NULL DEFAULT '',
			country VARCHAR(96) NOT NULL DEFAULT '',
			campus VARCHAR(191) NOT NULL DEFAULT '',
			package_name VARCHAR(191) NOT NULL DEFAULT '',
			pic_staff_id VARCHAR(64) NOT NULL DEFAULT '',
			status VARCHAR(96) NOT NULL DEFAULT '',
			progress INTEGER NOT NULL DEFAULT 0,
			last_schedule VARCHAR(96) NOT NULL DEFAULT '',
			current_stage VARCHAR(191) NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(64) PRIMARY KEY,
			code VARCHAR(96) NOT NULL UNIQUE,
			client_id VARCHAR(64) NOT NULL,
			package_name VARCHAR(191) NOT NULL DEFAULT '',
			total BIGINT NOT NULL DEFAULT 0,
			paid BIGINT NOT NULL DEFAULT 0,
			status VARCHAR(64) NOT NULL,
			due_date TIMESTAMPTZ NOT NULL,
			proof_note TEXT NOT NULL DEFAULT '',
			proof_file_name VARCHAR(255) NOT NULL DEFAULT '',
			proof_storage_path TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL,
			paid_at TIMESTAMPTZ
		)`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS proof_file_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE orders ADD COLUMN IF NOT EXISTS proof_storage_path TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS documents (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			name VARCHAR(191) NOT NULL,
			status VARCHAR(64) NOT NULL,
			reviewer VARCHAR(191) NOT NULL DEFAULT '',
			file_name VARCHAR(255) NOT NULL DEFAULT '',
			storage_path TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS storage_path TEXT NOT NULL DEFAULT ''`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS review_note TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS progress_stages (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			step INTEGER NOT NULL,
			title VARCHAR(191) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			status VARCHAR(64) NOT NULL,
			progress INTEGER NOT NULL DEFAULT 0,
			due_label VARCHAR(96) NOT NULL DEFAULT '',
			pic_name VARCHAR(191) NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS schedules (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			title VARCHAR(191) NOT NULL,
			date_label VARCHAR(96) NOT NULL DEFAULT '',
			time_label VARCHAR(96) NOT NULL DEFAULT '',
			location VARCHAR(191) NOT NULL DEFAULT '',
			status VARCHAR(64) NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(64) PRIMARY KEY,
			staff_id VARCHAR(64) NOT NULL,
			client_id VARCHAR(64) NOT NULL,
			time_label VARCHAR(32) NOT NULL DEFAULT '',
			title VARCHAR(191) NOT NULL,
			priority VARCHAR(64) NOT NULL DEFAULT '',
			status VARCHAR(64) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS expenses (
			id VARCHAR(64) PRIMARY KEY,
			staff_id VARCHAR(64) NOT NULL,
			client_id VARCHAR(64) NOT NULL,
			need VARCHAR(191) NOT NULL,
			category VARCHAR(96) NOT NULL DEFAULT '',
			amount BIGINT NOT NULL DEFAULT 0,
			status VARCHAR(64) NOT NULL,
			date_label VARCHAR(96) NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			receipt_file_name VARCHAR(255) NOT NULL DEFAULT '',
			receipt_storage_path TEXT NOT NULL DEFAULT ''
		)`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS receipt_file_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS receipt_storage_path TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS chat_conversations (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			staff_id VARCHAR(64) NOT NULL,
			last_message TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id VARCHAR(64) PRIMARY KEY,
			conversation_id VARCHAR(64) NOT NULL,
			sender_id VARCHAR(64) NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS pipeline_stages (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(191) NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,
			tone VARCHAR(32) NOT NULL DEFAULT 'mint',
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS service_packages (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(191) NOT NULL,
			category VARCHAR(96) NOT NULL DEFAULT '',
			description TEXT NOT NULL DEFAULT '',
			price BIGINT NOT NULL DEFAULT 0,
			price_is_from BOOLEAN NOT NULL DEFAULT FALSE,
			highlights TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL
		)`,
	}
}

func mysqlSchema() []string {
	return []string{
		`CREATE TABLE IF NOT EXISTS users (
			id VARCHAR(64) PRIMARY KEY,
			username VARCHAR(191) NOT NULL,
			name VARCHAR(191) NOT NULL,
			email VARCHAR(191) NOT NULL DEFAULT '',
			phone VARCHAR(64) NOT NULL DEFAULT '',
			role VARCHAR(32) NOT NULL,
			password_hash VARBINARY(255) NOT NULL,
			active BOOLEAN NOT NULL DEFAULT TRUE,
			created_at TIMESTAMP NOT NULL,
			UNIQUE KEY users_username_unique (username)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS clients (
			id VARCHAR(64) PRIMARY KEY,
			user_id VARCHAR(64) NULL,
			name VARCHAR(191) NOT NULL,
			email VARCHAR(191) NOT NULL DEFAULT '',
			phone VARCHAR(64) NOT NULL DEFAULT '',
			country VARCHAR(96) NOT NULL DEFAULT '',
			campus VARCHAR(191) NOT NULL DEFAULT '',
			package_name VARCHAR(191) NOT NULL DEFAULT '',
			pic_staff_id VARCHAR(64) NOT NULL DEFAULT '',
			status VARCHAR(96) NOT NULL DEFAULT '',
			progress INT NOT NULL DEFAULT 0,
			last_schedule VARCHAR(96) NOT NULL DEFAULT '',
			current_stage VARCHAR(191) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS orders (
			id VARCHAR(64) PRIMARY KEY,
			code VARCHAR(96) NOT NULL,
			client_id VARCHAR(64) NOT NULL,
			package_name VARCHAR(191) NOT NULL DEFAULT '',
			total BIGINT NOT NULL DEFAULT 0,
			paid BIGINT NOT NULL DEFAULT 0,
			status VARCHAR(64) NOT NULL,
			due_date TIMESTAMP NOT NULL,
			proof_note TEXT NOT NULL,
			proof_file_name VARCHAR(255) NOT NULL DEFAULT '',
			proof_storage_path TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL,
			paid_at TIMESTAMP NULL,
			UNIQUE KEY orders_code_unique (code)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS documents (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			name VARCHAR(191) NOT NULL,
			status VARCHAR(64) NOT NULL,
			reviewer VARCHAR(191) NOT NULL DEFAULT '',
			file_name VARCHAR(255) NOT NULL DEFAULT '',
			storage_path VARCHAR(1024) NOT NULL DEFAULT '',
			updated_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS file_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS storage_path VARCHAR(1024) NOT NULL DEFAULT ''`,
		`ALTER TABLE documents ADD COLUMN IF NOT EXISTS review_note TEXT NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS progress_stages (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			step INT NOT NULL,
			title VARCHAR(191) NOT NULL,
			description TEXT NOT NULL,
			status VARCHAR(64) NOT NULL,
			progress INT NOT NULL DEFAULT 0,
			due_label VARCHAR(96) NOT NULL DEFAULT '',
			pic_name VARCHAR(191) NOT NULL DEFAULT '',
			updated_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS schedules (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			title VARCHAR(191) NOT NULL,
			date_label VARCHAR(96) NOT NULL DEFAULT '',
			time_label VARCHAR(96) NOT NULL DEFAULT '',
			location VARCHAR(191) NOT NULL DEFAULT '',
			status VARCHAR(64) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS tasks (
			id VARCHAR(64) PRIMARY KEY,
			staff_id VARCHAR(64) NOT NULL,
			client_id VARCHAR(64) NOT NULL,
			time_label VARCHAR(32) NOT NULL DEFAULT '',
			title VARCHAR(191) NOT NULL,
			priority VARCHAR(64) NOT NULL DEFAULT '',
			status VARCHAR(64) NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS expenses (
			id VARCHAR(64) PRIMARY KEY,
			staff_id VARCHAR(64) NOT NULL,
			client_id VARCHAR(64) NOT NULL,
			need VARCHAR(191) NOT NULL,
			category VARCHAR(96) NOT NULL DEFAULT '',
			amount BIGINT NOT NULL DEFAULT 0,
			status VARCHAR(64) NOT NULL,
			date_label VARCHAR(96) NOT NULL DEFAULT '',
			description TEXT NOT NULL,
			receipt_file_name VARCHAR(255) NOT NULL DEFAULT '',
			receipt_storage_path VARCHAR(1024) NOT NULL DEFAULT ''
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS receipt_file_name VARCHAR(255) NOT NULL DEFAULT ''`,
		`ALTER TABLE expenses ADD COLUMN IF NOT EXISTS receipt_storage_path VARCHAR(1024) NOT NULL DEFAULT ''`,
		`CREATE TABLE IF NOT EXISTS chat_conversations (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			staff_id VARCHAR(64) NOT NULL,
			last_message TEXT NOT NULL,
			updated_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS chat_messages (
			id VARCHAR(64) PRIMARY KEY,
			conversation_id VARCHAR(64) NOT NULL,
			sender_id VARCHAR(64) NOT NULL,
			body TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS pipeline_stages (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(191) NOT NULL,
			position INT NOT NULL DEFAULT 0,
			tone VARCHAR(32) NOT NULL DEFAULT 'mint',
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS service_packages (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(191) NOT NULL,
			category VARCHAR(96) NOT NULL DEFAULT '',
			description TEXT NOT NULL,
			price BIGINT NOT NULL DEFAULT 0,
			price_is_from BOOLEAN NOT NULL DEFAULT FALSE,
			highlights TEXT NOT NULL,
			position INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
}
