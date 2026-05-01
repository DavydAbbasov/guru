CREATE USER products WITH PASSWORD 'products';
GRANT products TO postgres;
CREATE DATABASE products OWNER products;
