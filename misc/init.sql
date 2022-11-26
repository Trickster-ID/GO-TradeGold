
create table IF NOT EXISTS tbl_harga(
    id varchar(11) PRIMARY KEY,
    admin_id varchar(4),
    harga_topup NUMERIC,
    harga_buyback NUMERIC,
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_topup(
    id varchar(11) PRIMARY KEY,
    gram NUMERIC,
    harga NUMERIC,
    norek varchar(4),
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_rekening(
    id varchar(11) PRIMARY KEY,
    norek varchar(4),
    saldo NUMERIC,
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_transaksi(
    id varchar(11) PRIMARY KEY,
    gram NUMERIC,
    harga NUMERIC,
    norek varchar(4),
    created_date TIMESTAMP
);

CREATE OR REPLACE PROCEDURE sp_customer_topup(
	Pid VARCHAR(11),
	Pgram NUMERIC,
	Pharga NUMERIC,
	Pnorek VARCHAR(4)
)
AS $$
	DECLARE vLast numeric = (SELECT saldo FROM tbl_rekening WHERE norek = Pnorek order by created_date desc LIMIT 1);
BEGIN
    IF EXISTS (SELECT saldo FROM tbl_rekening WHERE norek = Pnorek order by created_date desc LIMIT 1) THEN
        INSERT INTO tbl_rekening VALUES
        (Pid, Pnorek, Pgram + vLast, CURRENT_TIMESTAMP);
    ELSE
        INSERT INTO tbl_rekening VALUES
        (Pid, Pnorek, Pgram, CURRENT_TIMESTAMP);
    END IF;
	
	INSERT INTO tbl_topup VALUES
	(Pid,Pgram,Pharga,Pnorek,CURRENT_TIMESTAMP);
END;
$$ LANGUAGE plpgsql;

-- CREATE OR REPLACE PROCEDURE sp_buyback(
-- 	Pid VARCHAR(11),
-- 	Pgram NUMERIC,
-- 	Pharga NUMERIC,
-- 	Pnorek VARCHAR(4)
-- )
-- AS $$
-- BEGIN
--     UPDATE tbl_rekening
--     SET saldo = saldo - Pgram, updated_date = CURRENT_TIMESTAMP
--     WHERE norek = Pnorek;
	
-- 	INSERT INTO tbl_transaksi VALUES
-- 	(Pid,Pgram,Pharga,Pnorek,CURRENT_TIMESTAMP);
-- END;
-- $$ LANGUAGE plpgsql;