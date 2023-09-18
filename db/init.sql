DROP TABLE IF EXISTS response;
DROP TABLE IF EXISTS request;

CREATE TABLE IF NOT EXISTS request
(
    request_id SERIAL NOT NULL PRIMARY KEY,
    request    text   NOT NULL
);

CREATE TABLE IF NOT EXISTS response
(
    response_id SERIAL                                 NOT NULL PRIMARY KEY,
    request_id  SERIAL REFERENCES request (request_id) NOT NULL,
    response    TEXT                                   NOT NULL
);