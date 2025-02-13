-- +goose Up
    create table if not exists users (
        user_id serial primary key,
        username text not null,
        email text not null
    );
    create type allowComments as enum (
        'yes',
        'no'
    );

    create table if not exists posts (
        post_id serial primary key,
        fk_user_id int not null,
        title text,
        body text,
        permission allowComments
    );

    create table if not exists comments(
        comment_id serial primary key,
        fk_user_id int not null,
        fk_post_id int not null,
        fk_parent_id int not null,
        body text,
        created_at timeStamp
    );

    alter table posts
    add constraint fk_posts_users
    foreign key (fk_user_id) references users(user_id);

    alter table comments
    add constraint fk_comments_users
    foreign key (fk_user_id) references users(user_id);

    alter table comments
    add constraint fk_comments_posts
    foreign key (fk_post_id) references users(user_id);

    alter table comments
    add constraint fk_parent_comment
    foreign key (fk_parent_id) references comments(comment_id);



-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
    drop table users;

    drop table posts;

    drop table comments;

    drop type allowComments;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
