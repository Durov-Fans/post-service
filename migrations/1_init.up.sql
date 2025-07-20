CREATE TABLE posts (
    UserId INTEGER   ,
    Id serial  PRIMARY KEY  UNIQUE,
    Title varchar(80),
    Media pg_catalog.json,
    CreatedAt DATE
);

CREATE TABLE IF NOT EXISTS comments
(
    id     INTEGER PRIMARY KEY,
    UserId INTEGER ,
    PostId INTEGER ,
    Description   TEXT NOT NULL,
    CreatedAt DATE,
    UpdatedAt DATE
);