
-- CREATE SCHEMA unitTest;
-- CREATE SCHEMA prduction;

-- create tables for unit test
CREATE TABLE customer (
    id INT PRIMARY KEY,
    username VARCHAR(255),
    addr VARCHAR(255),
    phone VARCHAR(53)
);

CREATE TABLE account (
  id INT PRIMARY KEY,
  balance FLOAT,
  FOREIGN KEY (id) REFERENCES unitTest.customer(id)
);

-- create tables for production code
-- CREATE TABLE production.customer (
--    id INT PRIMARY KEY,
--    username VARCHAR(255),
--    addr VARCHAR(255),
--    phone VARCHAR(53)
-- );

-- CREATE TABLE production.account (
--   id INT PRIMARY KEY,
--   balance FLOAT,
--   FOREIGN KEY (id) REFERENCES production.customer(id)
-- );