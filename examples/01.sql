CREATE TABLE public.accounts (
    id integer,
    name text
);

CREATE TABLE public.users (
    id integer,
    account_id integer,
    first_name text,
    last_name text,
    email text,
    title public.title_type,
    password text,
    password_expires_at date,
    tags text[],
    is_admin boolean
);
