
--
-- Name: voip_cdr; Type: TABLE; Schema: public; Owner: postgres; Tablespace:
--

CREATE TABLE cdr_import (
    id serial NOT NULL PRIMARY KEY,
    switch character varying(80) NOT NULL,
    cdr_source_type integer,
    callid character varying(80) NOT NULL,
    caller_id_number character varying(80) NOT NULL,
    caller_id_name character varying(80) NOT NULL,
    destination_number character varying(80) NOT NULL,
    dialcode character varying(10),
    state character varying(5),
    channel character varying(80),
    starting_date timestamp with time zone NOT NULL,
    duration integer NOT NULL,
    billsec integer NOT NULL,
    progresssec integer,
    answersec integer,
    waitsec integer,
    hangup_cause_id integer,
    hangup_cause character varying(80),
    direction integer,
    country_code character varying(3),
    accountcode character varying(40),
    buy_rate numeric(10,5),
    buy_cost numeric(12,5),
    sell_rate numeric(10,5),
    sell_cost numeric(12,5),
    data jsonb
);


--
-- cdr_source_type - type integer
-- acceptable values:
-- * unknown: 0
-- * freeswitch: 1
-- * asterisk: 2
-- * yate: 3
-- * kamailio: 4
-- * opensips: 5
--


--
-- direction - type integer
-- acceptable values:
-- * inbound: 1
-- * outbound: 2
--
