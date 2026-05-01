CREATE USER notification WITH PASSWORD 'notification';
GRANT notification TO postgres;
CREATE DATABASE notification OWNER notification;
