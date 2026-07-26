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
		`CREATE TABLE IF NOT EXISTS order_payments (
			id VARCHAR(64) PRIMARY KEY,
			order_id VARCHAR(64) NOT NULL,
			amount BIGINT NOT NULL DEFAULT 0,
			note TEXT NOT NULL DEFAULT '',
			proof_file_name VARCHAR(255) NOT NULL DEFAULT '',
			proof_storage_path TEXT NOT NULL DEFAULT '',
			status VARCHAR(64) NOT NULL,
			submitted_by VARCHAR(64) NOT NULL DEFAULT '',
			submitted_at TIMESTAMPTZ NOT NULL,
			verified_at TIMESTAMPTZ,
			reject_reason TEXT NOT NULL DEFAULT ''
		)`,
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
		`CREATE TABLE IF NOT EXISTS expense_categories (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(96) NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS shipments (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			staff_id VARCHAR(64) NOT NULL,
			direction VARCHAR(16) NOT NULL,
			courier VARCHAR(96) NOT NULL DEFAULT '',
			tracking_number VARCHAR(128) NOT NULL DEFAULT '',
			contents TEXT NOT NULL DEFAULT '',
			sender_address TEXT NOT NULL DEFAULT '',
			recipient_address TEXT NOT NULL DEFAULT '',
			status VARCHAR(32) NOT NULL,
			shipped_date_label VARCHAR(96) NOT NULL DEFAULT '',
			received_date_label VARCHAR(96) NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS shipment_couriers (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(96) NOT NULL UNIQUE,
			created_at TIMESTAMPTZ NOT NULL
		)`,
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
		`CREATE TABLE IF NOT EXISTS client_stage_history (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			stage_name VARCHAR(191) NOT NULL,
			entered_at TIMESTAMPTZ NOT NULL
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
		`CREATE TABLE IF NOT EXISTS text_templates (
			id VARCHAR(64) PRIMARY KEY,
			title VARCHAR(191) NOT NULL,
			body TEXT NOT NULL DEFAULT '',
			category VARCHAR(96) NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS institution_contacts (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(191) NOT NULL,
			category VARCHAR(96) NOT NULL DEFAULT '',
			phone VARCHAR(64) NOT NULL DEFAULT '',
			notes TEXT NOT NULL DEFAULT '',
			position INTEGER NOT NULL DEFAULT 0,
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS client_intake_forms (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL UNIQUE,
			email VARCHAR(191) NOT NULL DEFAULT '',
			full_name_en VARCHAR(191) NOT NULL DEFAULT '',
			gender VARCHAR(16) NOT NULL DEFAULT '',
			date_of_birth VARCHAR(32) NOT NULL DEFAULT '',
			place_of_birth VARCHAR(191) NOT NULL DEFAULT '',
			passport_number VARCHAR(64) NOT NULL DEFAULT '',
			phone_number VARCHAR(64) NOT NULL DEFAULT '',
			address TEXT NOT NULL DEFAULT '',
			postal_code VARCHAR(32) NOT NULL DEFAULT '',
			father_name VARCHAR(191) NOT NULL DEFAULT '',
			father_dob VARCHAR(32) NOT NULL DEFAULT '',
			father_phone VARCHAR(64) NOT NULL DEFAULT '',
			mother_name VARCHAR(191) NOT NULL DEFAULT '',
			mother_dob VARCHAR(32) NOT NULL DEFAULT '',
			mother_phone VARCHAR(64) NOT NULL DEFAULT '',
			school_name VARCHAR(191) NOT NULL DEFAULT '',
			school_location VARCHAR(191) NOT NULL DEFAULT '',
			dates_enrolled VARCHAR(64) NOT NULL DEFAULT '',
			dates_graduate VARCHAR(64) NOT NULL DEFAULT '',
			social_media_ig VARCHAR(191) NOT NULL DEFAULT '',
			submitted_at TIMESTAMPTZ NOT NULL,
			updated_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS activity_log (
			id VARCHAR(64) PRIMARY KEY,
			staff_id VARCHAR(64) NOT NULL,
			client_id VARCHAR(64) NOT NULL DEFAULT '',
			action_type VARCHAR(64) NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			created_at TIMESTAMPTZ NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS client_agreements (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL UNIQUE,
			agreement_version VARCHAR(32) NOT NULL DEFAULT 'v1',
			agreement_text TEXT NOT NULL DEFAULT '',
			full_name_typed VARCHAR(191) NOT NULL,
			agreed_at TIMESTAMPTZ NOT NULL,
			ip_address VARCHAR(64) NOT NULL DEFAULT '',
			user_agent VARCHAR(255) NOT NULL DEFAULT ''
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
		`CREATE TABLE IF NOT EXISTS order_payments (
			id VARCHAR(64) PRIMARY KEY,
			order_id VARCHAR(64) NOT NULL,
			amount BIGINT NOT NULL DEFAULT 0,
			note TEXT NOT NULL,
			proof_file_name VARCHAR(255) NOT NULL DEFAULT '',
			proof_storage_path VARCHAR(1024) NOT NULL DEFAULT '',
			status VARCHAR(64) NOT NULL,
			submitted_by VARCHAR(64) NOT NULL DEFAULT '',
			submitted_at TIMESTAMP NOT NULL,
			verified_at TIMESTAMP NULL,
			reject_reason TEXT NOT NULL
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
		`CREATE TABLE IF NOT EXISTS expense_categories (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(96) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			UNIQUE KEY expense_categories_name_unique (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS shipments (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			staff_id VARCHAR(64) NOT NULL,
			direction VARCHAR(16) NOT NULL,
			courier VARCHAR(96) NOT NULL DEFAULT '',
			tracking_number VARCHAR(128) NOT NULL DEFAULT '',
			contents TEXT NOT NULL,
			sender_address TEXT NOT NULL,
			recipient_address TEXT NOT NULL,
			status VARCHAR(32) NOT NULL,
			shipped_date_label VARCHAR(96) NOT NULL DEFAULT '',
			received_date_label VARCHAR(96) NOT NULL DEFAULT '',
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS shipment_couriers (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(96) NOT NULL,
			created_at TIMESTAMP NOT NULL,
			UNIQUE KEY shipment_couriers_name_unique (name)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
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
		`CREATE TABLE IF NOT EXISTS client_stage_history (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			stage_name VARCHAR(191) NOT NULL,
			entered_at TIMESTAMP NOT NULL
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
		`CREATE TABLE IF NOT EXISTS text_templates (
			id VARCHAR(64) PRIMARY KEY,
			title VARCHAR(191) NOT NULL,
			body TEXT NOT NULL,
			category VARCHAR(96) NOT NULL DEFAULT '',
			position INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS institution_contacts (
			id VARCHAR(64) PRIMARY KEY,
			name VARCHAR(191) NOT NULL,
			category VARCHAR(96) NOT NULL DEFAULT '',
			phone VARCHAR(64) NOT NULL DEFAULT '',
			notes TEXT NOT NULL,
			position INT NOT NULL DEFAULT 0,
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS client_intake_forms (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			email VARCHAR(191) NOT NULL DEFAULT '',
			full_name_en VARCHAR(191) NOT NULL DEFAULT '',
			gender VARCHAR(16) NOT NULL DEFAULT '',
			date_of_birth VARCHAR(32) NOT NULL DEFAULT '',
			place_of_birth VARCHAR(191) NOT NULL DEFAULT '',
			passport_number VARCHAR(64) NOT NULL DEFAULT '',
			phone_number VARCHAR(64) NOT NULL DEFAULT '',
			address TEXT NOT NULL,
			postal_code VARCHAR(32) NOT NULL DEFAULT '',
			father_name VARCHAR(191) NOT NULL DEFAULT '',
			father_dob VARCHAR(32) NOT NULL DEFAULT '',
			father_phone VARCHAR(64) NOT NULL DEFAULT '',
			mother_name VARCHAR(191) NOT NULL DEFAULT '',
			mother_dob VARCHAR(32) NOT NULL DEFAULT '',
			mother_phone VARCHAR(64) NOT NULL DEFAULT '',
			school_name VARCHAR(191) NOT NULL DEFAULT '',
			school_location VARCHAR(191) NOT NULL DEFAULT '',
			dates_enrolled VARCHAR(64) NOT NULL DEFAULT '',
			dates_graduate VARCHAR(64) NOT NULL DEFAULT '',
			social_media_ig VARCHAR(191) NOT NULL DEFAULT '',
			submitted_at TIMESTAMP NOT NULL,
			updated_at TIMESTAMP NOT NULL,
			UNIQUE KEY client_intake_forms_client_unique (client_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS activity_log (
			id VARCHAR(64) PRIMARY KEY,
			staff_id VARCHAR(64) NOT NULL,
			client_id VARCHAR(64) NOT NULL DEFAULT '',
			action_type VARCHAR(64) NOT NULL,
			description TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
		`CREATE TABLE IF NOT EXISTS client_agreements (
			id VARCHAR(64) PRIMARY KEY,
			client_id VARCHAR(64) NOT NULL,
			agreement_version VARCHAR(32) NOT NULL DEFAULT 'v1',
			agreement_text TEXT NOT NULL,
			full_name_typed VARCHAR(191) NOT NULL,
			agreed_at TIMESTAMP NOT NULL,
			ip_address VARCHAR(64) NOT NULL DEFAULT '',
			user_agent VARCHAR(255) NOT NULL DEFAULT '',
			UNIQUE KEY client_agreements_client_unique (client_id)
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci`,
	}
}
