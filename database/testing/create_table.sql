CREATE TABLE pages (
                     url VARCHAR(256) PRIMARY KEY,
                     parent VARCHAR(256) NULL,
                     domain VARCHAR(256) NOT NULL,
                     timestamp TIMESTAMP NULL,
                     body TEXT NULL
);