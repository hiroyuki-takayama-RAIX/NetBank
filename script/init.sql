-- Create database
-- CREATE DATABASE netbank;

-- Connect to the "netbank" database
-- \c netbank;

-- Create schema
-- CREATE SCHEMA information;

-- Create role
-- CREATE ROLE hoge WITH LOGIN PASSWORD 'passw0rd';

-- Grant privileges to "hoge" on the "information" schema
-- GRANT ALL PRIVILEGES ON SCHEMA information TO hoge;

-- Create the "customer" table
CREATE TABLE customer (
    id INT PRIMARY KEY,
    username VARCHAR(255),
    addr VARCHAR(255),
    phone VARCHAR(53)
);

-- INSERT INTO customer (id, username, addr, phone) VALUES (1001, 'John', 'Los Angeles, California', '(213) 555 0147');

-- Create the "account" table
CREATE TABLE account (
  id INT PRIMARY KEY,
  balance FLOAT,
  FOREIGN KEY (id) REFERENCES customer(id)
);

-- Insert sample records
-- INSERT INTO account (id, balance) VALUES (1001, 0);

-- add privileges on hoge
-- GRANT ALL PRIVILEGES ON information.account TO hoge;
-- GRANT ALL PRIVILEGES ON information.customer TO hoge;