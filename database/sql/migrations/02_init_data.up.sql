INSERT INTO invite_codes(code)
VALUES ('code');

INSERT INTO posts(content)
VALUES ('{
  "data": "First post"
}');

INSERT INTO posts(content)
VALUES ('{
  "data": "Second post"
}');

UPDATE posts
SET next_post_id = 2
WHERE id = 1;