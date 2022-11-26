
create table IF NOT EXISTS tbl_harga(
    id varchar(11),
    admin_id varchar(4),
    harga_topup float,
    harga_buyback float,
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_topup(
    id varchar(11),
    gram decimal,
    harga float,
    norek varchar(4),
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_rekening(
    id varchar(11),
    norek varchar(4),
    saldo float,
    created_date TIMESTAMP
);

create table IF NOT EXISTS tbl_transaksi(
    id varchar(11),
    norek varchar(4),
    saldo float,
    created_date TIMESTAMP
);