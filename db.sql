DROP TABLE IF EXISTS b_tx;
CREATE TABLE b_tx
(
    id bigserial,
    asset varchar(20),
    fromAddr varchar(66),
    toAddr varchar(66),
    value numeric(28,8),
    status int default 0,
    blockNum bigint default 0,
    txhash varchar(66),
    ext1 varchar(200),
    blockTime timestamp
    without time zone,
    createTime timestamp without time zone,
    updateTime timestamp without time zone,
    constraint uk_b_tx_txhash unique(txhash),
    PRIMARY KEY(id)
);