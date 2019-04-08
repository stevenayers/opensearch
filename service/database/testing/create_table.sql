CREATE TABLE pages (
                     url VARCHAR(256) PRIMARY KEY,
                     parent VARCHAR(256) NULL,
                     timestamp TIMESTAMP NULL,
                     body TEXT NULL
);