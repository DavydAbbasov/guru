SELECT 'CREATE DATABASE notification'
WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'notification')\gexec
