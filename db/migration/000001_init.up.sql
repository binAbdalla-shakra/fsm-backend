-- Enable PostgreSQL extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS postgis;

-- 1. Permissions Registry
CREATE TABLE IF NOT EXISTS permissions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(100) UNIQUE NOT NULL, -- e.g., 'ticket:create', 'tech:telemetry'
    name VARCHAR(150) NOT NULL,
    description TEXT,
    group_name VARCHAR(100) NOT NULL, -- e.g., 'Ticket Management', 'System Config'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Roles Hierarchy
CREATE TABLE IF NOT EXISTS roles (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    hierarchy_level INTEGER NOT NULL DEFAULT 1, -- higher values indicate higher privilege
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

-- 3. Role-Permissions Join
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    permission_id UUID REFERENCES permissions(id) ON DELETE CASCADE,
    PRIMARY KEY (role_id, permission_id)
);

-- 4. Core Users Table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(30) UNIQUE NOT NULL,
    email VARCHAR(150) UNIQUE,
    password_hash VARCHAR(255), -- for potential future password authentication
    status VARCHAR(30) NOT NULL DEFAULT 'PENDING_VERIFICATION', -- PENDING, ACTIVE, SUSPENDED
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status) WHERE deleted_at IS NULL;

-- 5. User-Roles Join
CREATE TABLE IF NOT EXISTS user_roles (
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- 6. User Active Sessions (Device Management)
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    device_name VARCHAR(150),
    refresh_token_hash VARCHAR(255) NOT NULL,
    ip_address VARCHAR(45),
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_sessions_user_device ON user_sessions(user_id, device_id);
CREATE INDEX IF NOT EXISTS idx_sessions_refresh ON user_sessions(refresh_token_hash);

-- 7. Customer Profile
CREATE TABLE IF NOT EXISTS customers (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    account_number VARCHAR(100) UNIQUE NOT NULL,
    plan_type VARCHAR(50) NOT NULL DEFAULT 'STANDARD', -- PREMIUM, ADVANCED, STANDARD
    current_speed VARCHAR(50) NOT NULL DEFAULT '100 Mbps',
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX IF NOT EXISTS idx_customers_account ON customers(account_number) WHERE deleted_at IS NULL;

-- 8. Technician Profile
CREATE TABLE IF NOT EXISTS technicians (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(30) NOT NULL DEFAULT 'OFFLINE', -- ONLINE, OFFLINE, ON_TRIP, BUSY
    workload INTEGER NOT NULL DEFAULT 0,
    skills TEXT[] NOT NULL DEFAULT '{}',
    location GEOMETRY(Point, 4326),
    zone_assignment VARCHAR(150),
    rating NUMERIC(3, 2) DEFAULT 5.0,
    tasks_completed INTEGER NOT NULL DEFAULT 0,
    shift_start TIME,
    shift_end TIME,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX IF NOT EXISTS idx_technicians_location ON technicians USING gist(location);
CREATE INDEX IF NOT EXISTS idx_technicians_skills ON technicians USING gin(skills);
CREATE INDEX IF NOT EXISTS idx_technicians_status ON technicians(status) WHERE deleted_at IS NULL;

-- 9. Tickets Table (Reporting & Verification Metadata)
CREATE TABLE IF NOT EXISTS tickets (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_number VARCHAR(50) UNIQUE NOT NULL, -- e.g., 'TK-88394'
    customer_id UUID NOT NULL REFERENCES customers(id) ON DELETE RESTRICT,
    technician_id UUID REFERENCES technicians(id) ON DELETE SET NULL,
    title VARCHAR(200) NOT NULL,
    description TEXT NOT NULL,
    category VARCHAR(100) NOT NULL, -- No Internet, Slow Connection, Intermittent, Equipment
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM', -- LOW, MEDIUM, HIGH, URGENT
    status VARCHAR(40) NOT NULL DEFAULT 'REPORTED', -- REPORTED, AUTO_DISPATCHING, DISPATCHED, IN_PROGRESS, RESOLVED, COMPLETED
    landmark TEXT,
    location GEOMETRY(Point, 4326) NOT NULL,
    before_photo_url VARCHAR(255),
    after_photo_url VARCHAR(255),
    otp_code VARCHAR(6) NOT NULL, -- Secure Verification OTP
    rating_score INTEGER,
    rating_tags VARCHAR(100)[],
    rating_comment TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE,
    created_by UUID,
    updated_by UUID,
    deleted_by UUID
);

CREATE INDEX IF NOT EXISTS idx_tickets_number ON tickets(ticket_number);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tickets_location ON tickets USING gist(location);
CREATE INDEX IF NOT EXISTS idx_tickets_tech ON tickets(technician_id) WHERE deleted_at IS NULL;

-- 10. Ticket Progress Logs Audit
CREATE TABLE IF NOT EXISTS ticket_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_id UUID NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    old_status VARCHAR(40),
    new_status VARCHAR(40) NOT NULL,
    action VARCHAR(100) NOT NULL,
    notes TEXT,
    performed_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_logs_ticket ON ticket_logs(ticket_id);

-- 11. OTP Verification Registry
CREATE TABLE IF NOT EXISTS otp_verifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    phone VARCHAR(30) NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    action_type VARCHAR(50) NOT NULL, -- SIGNUP, LOGIN, TICKET_COMPLETE
    is_verified BOOLEAN NOT NULL DEFAULT FALSE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_otp_phone_action ON otp_verifications(phone, action_type);

-- 12. System Notifications
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(150) NOT NULL,
    body TEXT NOT NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    channel VARCHAR(30) NOT NULL DEFAULT 'PUSH', -- PUSH, SMS, EMAIL
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id);

-- 13. Data Modification Audits
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- INSERT, UPDATE, DELETE
    table_name VARCHAR(100) NOT NULL,
    record_id UUID NOT NULL,
    old_values JSONB,
    new_values JSONB,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_audit_table_record ON audit_logs(table_name, record_id);
