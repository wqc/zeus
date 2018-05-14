create database if not exists zeus;

use zeus;

CREATE TABLE IF NOT EXISTS huobi (
	ts					int(11)	not null,
	coin_pair			varchar(128) not null,
	price				int(11) not null,
	time_stamp			timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
	PRIMARY KEY			(ts, coin_pair)
);

create table if not exists binance (
	ts					int(11) not null,
	coin_pair			varchar(128) not null,
);

create table if not exists coin_history (
	ts					int(11) not null,
	exchange			varchar(128) not null,
	coin_pair			varchar(128) not null,
	rt_price			float not null,
	trading_volume		float not null,
	time_stamp			timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
);
