-- +goose Up
ALTER TABLE users ADD COLUMN user_img TEXT NOT NULL DEFAULT 'https://i.ibb.co/5ntJp5C/defaultrobot2.png';

-- +goose Down
ALTER TABLE users DROP COLUMN user_img;