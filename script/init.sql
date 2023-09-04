-- Create database
CREATE DATABASE netbank;

-- Connect to the "netbank" database
\c netbank;

-- Create schema
CREATE SCHEMA information;

-- Create role
CREATE ROLE hoge WITH LOGIN PASSWORD 'passw0rd';

-- Grant privileges to "hoge" on the "information" schema
GRANT ALL PRIVILEGES ON SCHEMA information TO hoge;

-- Create the "customer" table
CREATE TABLE information.customer (
    id INT PRIMARY KEY,
    username VARCHAR(255),
    addr VARCHAR(255),
    phone VARCHAR(53)
);

INSERT INTO information.customer (id, username, addr, phone) VALUES (1001, 'John', 'Los Angeles, California', '(213) 555 0147');

-- Create the "account" table
CREATE TABLE information.account (
  id INT PRIMARY KEY,
  balance FLOAT,
  customer_id INT REFERENCES information.customer(id)
);

-- Insert sample records
INSERT INTO information.account (id, balance, customer_id) VALUES (1001, 0, 1001);

-- add privileges on hoge
GRANT ALL PRIVILEGES ON information.account TO hoge;
GRANT ALL PRIVILEGES ON information.customer TO hoge;