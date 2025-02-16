-- +goose Up
    create table if not exists users (
        id serial primary key,
        username text not null,
        email text not null unique
    );

    create table if not exists posts (
        id serial primary key,
        user_id int not null,
        title text not null,
        body text not null,
        permission bool,
        foreign key (user_id) references users(id) on delete cascade
    );

    create table if not exists comments(
        id serial primary key,
        user_id int not null,
        post_id int not null,
        parent_id int,
        body text not null,
        created_at timeStamp default current_timestamp,
        foreign key (user_id) references users(id) on delete cascade,
        foreign key (post_id) references posts(id) on delete cascade,
        foreign key (parent_id) references comments(id) on delete cascade
    );



-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down

    drop table comments;

    drop table posts;

    drop table users;

-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
