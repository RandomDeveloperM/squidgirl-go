create database squidgirl default character set utf8;

/* ログインユーザー情報 */
create table users
(
    id int not null unique auto_increment,
    name varchar(256) not null,
    passhash varchar(256) not null,
    permission int not null,
    created_at datetime not null,
    updated_at datetime not null,
    primary key (id)
) engine=innodb;

/* アーカイブファイル情報 */
create table books
(
    id int not null unique auto_increment,
    hash varchar(64) not null,
    file_path varchar(1024) not null,
    file_size int not null,
    page int not null,
    created_at datetime not null,
    updated_at datetime not null,
    primary key (id)
) engine=innodb;

/* アーカイブの表示情報 */
create table histoires
(
    id int not null unique auto_increment,
    user_name varchar(256) not null,
    book_hash varchar(64) not null,
    read_pos int not null,
    reaction int not null,
    created_at datetime not null,
    updated_at datetime not null,
    primary key (id)
) engine=innodb;

