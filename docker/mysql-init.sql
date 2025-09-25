
USE test_db;

CREATE TABLE users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    username VARCHAR(100) UNIQUE,
    email_address VARCHAR(255) UNIQUE,
    phone_number VARCHAR(20),
    date_of_birth DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);

CREATE TABLE customers (
    customer_id INT AUTO_INCREMENT PRIMARY KEY,
    full_name VARCHAR(100),
    email VARCHAR(255),
    phone VARCHAR(20),
    credit_card_number VARCHAR(20),
    ssn VARCHAR(11),
    driver_license VARCHAR(20),
    address TEXT,
    postal_code VARCHAR(10),
    registration_date DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- Employee information
CREATE TABLE employees (
    emp_id INT AUTO_INCREMENT PRIMARY KEY,
    employee_number VARCHAR(20) UNIQUE,
    first_name VARCHAR(50),
    last_name VARCHAR(50),
    work_email VARCHAR(255),
    personal_email VARCHAR(255),
    home_phone VARCHAR(20),
    mobile_phone VARCHAR(20),
    passport_number VARCHAR(20),
    national_id VARCHAR(20),
    bank_account VARCHAR(30),
    department VARCHAR(50),
    position VARCHAR(100),
    salary DECIMAL(10,2),
    hire_date DATE
);

CREATE TABLE access_logs (
    log_id BIGINT AUTO_INCREMENT PRIMARY KEY,
    user_id INT,
    ip_address VARCHAR(45),
    mac_address VARCHAR(17),
    session_id VARCHAR(100),
    login_time TIMESTAMP,
    logout_time TIMESTAMP,
    status VARCHAR(20),
    INDEX idx_user_id (user_id),
    INDEX idx_ip_address (ip_address)
);

CREATE TABLE orders (
    order_id INT AUTO_INCREMENT PRIMARY KEY,
    customer_id INT,
    account_number VARCHAR(20),
    payment_method VARCHAR(50),
    total_amount DECIMAL(10,2),
    order_date DATE,
    shipping_address TEXT,
    billing_zip_code VARCHAR(10),
    status VARCHAR(30)
);

INSERT INTO users (first_name, last_name, username, email_address, phone_number, date_of_birth) VALUES
('John', 'Doe', 'johndoe', 'john.doe@example.com', '+1-555-0101', '1990-01-15'),
('Jane', 'Smith', 'janesmith', 'jane.smith@example.com', '+1-555-0102', '1985-03-22'),
('Bob', 'Johnson', 'bobjohnson', 'bob.johnson@example.com', '+1-555-0103', '1992-07-08');

INSERT INTO customers (full_name, email, phone, credit_card_number, ssn, driver_license, address, postal_code) VALUES
('Alice Williams', 'alice.w@example.com', '+1-555-0201', '4111-1111-1111-1111', '123-45-6789', 'D12345678', '123 Main St, Anytown', '12345'),
('Charlie Brown', 'charlie.b@example.com', '+1-555-0202', '5555-5555-5555-4444', '987-65-4321', 'D87654321', '456 Oak Ave, Somewhere', '54321');

INSERT INTO employees (employee_number, first_name, last_name, work_email, personal_email, home_phone, mobile_phone, passport_number, national_id, bank_account, department, position, salary, hire_date) VALUES
('EMP001', 'Sarah', 'Connor', 'sarah.connor@company.com', 'sarah.personal@email.com', '+1-555-0301', '+1-555-0302', 'P123456789', 'N987654321', 'ACC123456789', 'IT', 'Software Engineer', 75000.00, '2020-01-15'),
('EMP002', 'Michael', 'Scott', 'michael.scott@company.com', 'michael.personal@email.com', '+1-555-0303', '+1-555-0304', 'P987654321', 'N123456789', 'ACC987654321', 'Management', 'Regional Manager', 95000.00, '2019-03-01');

INSERT INTO access_logs (user_id, ip_address, mac_address, session_id, login_time, status) VALUES
(1, '192.168.1.100', '00:11:22:33:44:55', 'sess_abc123', '2024-01-15 09:00:00', 'active'),
(2, '10.0.0.50', '66:77:88:99:AA:BB', 'sess_def456', '2024-01-15 09:15:00', 'active'),
(3, '172.16.0.25', 'CC:DD:EE:FF:11:22', 'sess_ghi789', '2024-01-15 09:30:00', 'logged_out');

INSERT INTO orders (customer_id, account_number, payment_method, total_amount, order_date, shipping_address, billing_zip_code, status) VALUES
(1, 'ACC001122334455', 'Credit Card', 299.99, '2024-01-10', '123 Main St, Anytown', '12345', 'shipped'),
(2, 'ACC556677889900', 'Bank Transfer', 149.50, '2024-01-12', '456 Oak Ave, Somewhere', '54321', 'processing');

CREATE DATABASE IF NOT EXISTS hr_system;
USE hr_system;

CREATE TABLE staff (
    id INT AUTO_INCREMENT PRIMARY KEY,
    employee_id VARCHAR(20),
    fname VARCHAR(50),
    lname VARCHAR(50),
    useremail VARCHAR(255),
    social_security_number VARCHAR(11),
    phone_num VARCHAR(20),
    home_address TEXT,
    zipcode VARCHAR(10)
);

CREATE DATABASE IF NOT EXISTS inventory_db;
USE inventory_db;

CREATE TABLE products (
    product_id INT AUTO_INCREMENT PRIMARY KEY,
    product_name VARCHAR(100),
    sku VARCHAR(50),
    price DECIMAL(10,2),
    stock_quantity INT,
    supplier_contact_email VARCHAR(255),
    created_date DATE
);

CREATE TABLE suppliers (
    supplier_id INT AUTO_INCREMENT PRIMARY KEY,
    company_name VARCHAR(100),
    contact_person VARCHAR(100),
    email_addr VARCHAR(255),
    phone_number VARCHAR(20),
    business_address TEXT,
    zip_code VARCHAR(10)
);

CREATE DATABASE IF NOT EXISTS classifier_meta;
USE classifier_meta;

CREATE TABLE IF NOT EXISTS database_connections (
    id CHAR(36) PRIMARY KEY,
    host VARCHAR(255) NOT NULL,
    port INT NOT NULL,
    username VARCHAR(128) NOT NULL,
    encrypted_password TEXT NOT NULL,
    database_name VARCHAR(255),
    description TEXT,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL,
    last_scanned_at DATETIME(6) NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1
);

CREATE TABLE IF NOT EXISTS scan_results (
    id CHAR(36) PRIMARY KEY,
    database_id CHAR(36) NOT NULL,
    started_at DATETIME(6) NOT NULL,
    completed_at DATETIME(6) NULL,
    status VARCHAR(32) NOT NULL,
    error_message TEXT,
    schemas_json LONGTEXT NULL,
    summary_json LONGTEXT NULL,
    INDEX idx_scan_database (database_id),
    INDEX idx_scan_status (status),
    INDEX idx_scan_started_at (started_at)
);

CREATE TABLE IF NOT EXISTS classification_patterns (
    id CHAR(36) PRIMARY KEY,
    information_type VARCHAR(64) NOT NULL,
    pattern VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    priority INT NOT NULL,
    is_active TINYINT(1) NOT NULL DEFAULT 1,
    created_at DATETIME(6) NOT NULL,
    updated_at DATETIME(6) NOT NULL
);

CREATE USER IF NOT EXISTS 'metauser'@'%' IDENTIFIED BY 'metapass';
GRANT ALL PRIVILEGES ON classifier_meta.* TO 'metauser'@'%';
FLUSH PRIVILEGES;
