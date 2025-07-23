CREATE TYPE subLevel AS enum (
    'None',
    'Supporter',
    'Premium',
    'Exclusive');
CREATE TABLE posts (
    UserId INTEGER   ,
    Id serial  PRIMARY KEY  UNIQUE,
    Description TEXT,
    Media jsonb,
    CreatedAt DATE,
    likeNum integer default 0,
    Paid bool default false,
    SubLevel subLevel default 'None'
);
CREATE TABLE comments
(
    Id serial  PRIMARY KEY  UNIQUE,
    UserId INTEGER ,
    PostId INTEGER ,
    Description   TEXT NOT NULL,
    CreatedAt DATE default now(),
    UpdatedAt DATE default now()
);
CREATE TABLE likes
(
    Id serial  PRIMARY KEY  UNIQUE,
    UserId INTEGER ,
    PostId INTEGER
);
