create database if not exists zeus;

use zeus;

CREATE TABLE IF NOT EXISTS market (
	id					bigint auto_increment primary key,
	symbol				varchar(256) not null,
	exchange			varchar(256) not null,
	ptype               int,
	ts					bigint,
	price				float,
	quantity			float,
	update_time			timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
