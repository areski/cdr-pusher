-- .dump cdr
PRAGMA foreign_keys=OFF;
BEGIN TRANSACTION;
CREATE TABLE cdr (
    caller_id_name VARCHAR,
    caller_id_number VARCHAR,
    destination_number VARCHAR,
    context VARCHAR,
    start_stamp DATETIME,
    answer_stamp DATETIME,
    end_stamp DATETIME,
    duration INTEGER,
    billsec INTEGER,
    hangup_cause VARCHAR,
    hangup_cause_q850 INTEGER,
    uuid VARCHAR,
    bleg_uuid VARCHAR,
    account_code VARCHAR
);

INSERT INTO "cdr" VALUES('Outbound Call','800123123','34650881188','default','2015-01-14 17:58:01','2015-01-14 17:58:01','2015-01-14 17:59:56',50,50,'NORMAL_CLEARING',16,'2bbe83f7-5111-4b5b-9626-c5154608d4ee','','');
INSERT INTO "cdr" VALUES('Outbound Call','800123123','34650881100','default','2015-01-14 17:59:02','2015-01-14 17:59:02','2015-01-14 17:59:17',15,15,'NORMAL_CLEARING',16,'2bbe83f7-5111-4b5b-1324-joij32j2o344','','');
COMMIT;
