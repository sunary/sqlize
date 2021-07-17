package main

import (
	"github.com/sunary/sqlize/sql-parser"
)

func main() {
	sql := `
ALTER TABLE distributors ADD COLUMN address varchar(30);

ALTER TABLE distributors DROP COLUMN address RESTRICT;

ALTER TABLE distributors
    ALTER COLUMN address TYPE varchar(80),
    ALTER COLUMN name TYPE varchar(100);

ALTER TABLE foo
    ALTER COLUMN foo_timestamp SET DATA TYPE timestamp with time zone
    USING
        timestamp with time zone 'epoch' + foo_timestamp * interval '1 second';

ALTER TABLE distributors RENAME COLUMN address TO city;


ALTER TABLE foo
    ALTER COLUMN foo_timestamp SET DATA TYPE timestamp with time zone
    USING
        timestamp with time zone 'epoch' + foo_timestamp * interval '1 second';


ALTER TABLE foo
    ALTER COLUMN foo_timestamp DROP DEFAULT,
    ALTER COLUMN foo_timestamp TYPE timestamp with time zone
    USING
        timestamp with time zone 'epoch' + foo_timestamp * interval '1 second',
    ALTER COLUMN foo_timestamp SET DEFAULT now();

ALTER TABLE distributors RENAME COLUMN address TO city;


ALTER TABLE distributors RENAME TO suppliers;


ALTER TABLE distributors ALTER COLUMN street DROP NOT NULL;

ALTER TABLE distributors DROP CONSTRAINT zipchk;

ALTER TABLE distributors ADD CONSTRAINT distfk FOREIGN KEY (address) REFERENCES addresses (address) MATCH FULL;

ALTER TABLE distributors ADD PRIMARY KEY (dist_id);


CREATE TABLE accounts (
    id serial PRIMARY KEY ,
    user_pseudo_id VARCHAR(50) NOT NULL,
    cake_user_id VARCHAR(250) NOT NULL ,
    event_timestamp TIMESTAMP NOT NULL ,
    imported_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
    UNIQUE (cake_user_id, imported_at) 
);
`
	p := sql_parser.NewParser(true, false)
	err := p.Parser(sql)
	if err != nil {
		println(err.Error())
	}
}
