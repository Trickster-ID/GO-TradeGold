
create table IF NOT EXISTS tbl_harga(
    id varchar(11) PRIMARY KEY,
    admin_id varchar(4),
    harga_topup NUMERIC,
    harga_buyback NUMERIC,
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_rekening(
    norek varchar(4) PRIMARY KEY,
    saldo NUMERIC,
    created_date TIMESTAMP,
    update_date TIMESTAMP
);

create table IF NOT EXISTS tbl_topup(
    id varchar(11) PRIMARY KEY,
    gram NUMERIC,
    harga_topup NUMERIC,
    harga_buyback NUMERIC,
    norek varchar(4),
    saldo NUMERIC,
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_transaksi(
    id varchar(11) PRIMARY KEY,
    gram NUMERIC,
    harga_topup NUMERIC,
    harga_buyback NUMERIC,
    norek varchar(4),
    saldo NUMERIC,
    created_date TIMESTAMP
);

CREATE OR REPLACE PROCEDURE sp_customer_topup(
	Pid VARCHAR(11),
	Pgram NUMERIC,
	Pnorek VARCHAR(4)
)
AS $$
	DECLARE vLast numeric = (SELECT saldo FROM tbl_rekening WHERE norek = Pnorek order by created_date desc LIMIT 1);
BEGIN
    IF EXISTS (SELECT saldo FROM tbl_rekening WHERE norek = Pnorek order by created_date desc LIMIT 1) THEN
        UPDATE tbl_rekening
        SET saldo = vLast + Pgram, update_date = CURRENT_TIMESTAMP
        WHERE norek = Pnorek;

        INSERT INTO tbl_topup 
        SELECT Pid,Pgram,harga_topup, harga_buyback,Pnorek,vLast + Pgram,CURRENT_TIMESTAMP 
        FROM tbl_harga ORDER BY created_date DESC LIMIT 1;
    ELSE
        INSERT INTO tbl_rekening VALUES
        (Pnorek, Pgram, CURRENT_TIMESTAMP);

        INSERT INTO tbl_topup 
        SELECT Pid,Pgram,harga_topup, harga_buyback,Pnorek,Pgram,CURRENT_TIMESTAMP 
        FROM tbl_harga ORDER BY created_date DESC LIMIT 1;
    END IF;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE PROCEDURE sp_buyback(
	Pid VARCHAR(11),
	Pgram NUMERIC,
	Pnorek VARCHAR(4)
)
AS $$
	DECLARE vLast numeric = (SELECT saldo FROM tbl_rekening WHERE norek = Pnorek order by created_date desc LIMIT 1);
BEGIN
    UPDATE tbl_rekening
    SET saldo = vLast - Pgram, update_date = CURRENT_TIMESTAMP
    WHERE norek = Pnorek;

    INSERT INTO tbl_transaksi 
    SELECT Pid,Pgram,harga_topup, harga_buyback,Pnorek,vLast - Pgram,CURRENT_TIMESTAMP 
    FROM tbl_harga ORDER BY created_date DESC LIMIT 1;
END;
$$ LANGUAGE plpgsql;