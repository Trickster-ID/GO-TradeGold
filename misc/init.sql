
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
    norek varchar(4) PRIMARY KEY,
    saldo NUMERIC,
    created_date TIMESTAMP,
    updated_date TIMESTAMP
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
BEGIN
	IF EXISTS (SELECT * FROM tbl_rekening WHERE norek = Pnorek ) THEN
		UPDATE tbl_rekening
		SET saldo = saldo + Pgram, updated_date = CURRENT_TIMESTAMP
		WHERE norek = Pnorek;
	ELSE
		INSERT INTO tbl_rekening VALUES
		(Pnorek, Pgram, CURRENT_TIMESTAMP);
	END IF;
	
	INSERT INTO tbl_topup VALUES
	(Pid,Pgram,Pharga,Pnorek,CURRENT_TIMESTAMP);
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE PROCEDURE sp_buyback(
	Pid VARCHAR(11),
	Pgram NUMERIC,
	Pharga NUMERIC,
	Pnorek VARCHAR(4)
)
AS $$
BEGIN
    UPDATE tbl_rekening
    SET saldo = saldo - Pgram, updated_date = CURRENT_TIMESTAMP
    WHERE norek = Pnorek;
	
	INSERT INTO tbl_transaksi VALUES
	(Pid,Pgram,Pharga,Pnorek,CURRENT_TIMESTAMP);
END;
$$ LANGUAGE plpgsql;