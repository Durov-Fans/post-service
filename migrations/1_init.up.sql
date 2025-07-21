CREATE TYPE subLevel AS enum (
    'None',
    'Supporter',
    'Premium',
    'Exclusive');
CREATE TABLE posts (
    UserId INTEGER   ,
    Id serial  PRIMARY KEY  UNIQUE,
    Title varchar(80),
    Media jsonb,
    CreatedAt DATE,
    Paid bool default false,
    SubLevel subLevel default 'None'
);
CREATE TABLE comments
(
    id     INTEGER PRIMARY KEY,
    UserId INTEGER ,
    PostId INTEGER ,
    Description   TEXT NOT NULL,
    CreatedAt DATE,
    UpdatedAt DATE
);