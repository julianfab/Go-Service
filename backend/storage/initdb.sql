CREATE USER IF NOT EXISTS julian;
CREATE DATABASE dbprueba;
GRANT ALL ON DATABASE dbprueba TO julian;



DROP SEQUENCE IF EXISTS dominio_seq CASCADE;
CREATE SEQUENCE dominio_seq;

DROP TABLE IF EXISTS dominio CASCADE;

CREATE TABLE dominio(
  id INTEGER DEFAULT nextval('dominio_seq') NOT NULL PRIMARY KEY,
  name VARCHAR(50) NOT NULL,
  servers_changed BOOLEAN DEFAULT false,
  ssl_grade VARCHAR(3) NOT NULL,
  previous_ssl_grade VARCHAR(3),
  logo VARCHAR(500) NOT NULL,
  title VARCHAR(50) NOT NULL,
  is_down BOOLEAN NOT NULL
);

DROP SEQUENCE IF EXISTS server_seq CASCADE;
CREATE SEQUENCE server_seq;

DROP TABLE IF EXISTS server CASCADE;

CREATE TABLE server(
  id INTEGER DEFAULT nextval('server_seq') NOT NULL PRIMARY KEY,
  address VARCHAR(70) NOT NULL,
  ssl_grade VARCHAR(3) NOT NULL,
  country VARCHAR(3) NOT NULL,
  owner VARCHAR(50) NOT NULL,

  id_dominio INTEGER NOT NULL,
  CONSTRAINT server_fk FOREIGN KEY (id_dominio) REFERENCES dominio(id)

);
-- cockroach cert create-client julian --certs-dir=certs --ca-key=my-safe-directory/ca.key
