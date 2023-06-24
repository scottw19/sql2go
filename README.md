sql2go takes a --schema-only pg_dump file and creates Go structs for all the tables.

Flags:
```
  -i string
      Input filepath
  -j  Include json struct tags
  -o string
      Output filepath
```

Example:
```
./sql2go -i examples/01.sql -j
type Account struct {
  Id   int    `json:"id"`
  Name string `json:"name"`
}

type User struct {
  Id                int      `json:"id"`
  AccountId         int      `json:"account_id"`
  FirstName         string   `json:"first_name"`
  LastName          string   `json:"last_name"`
  Email             string   `json:"email"`
  Title             string   `json:"title"`
  Password          string   `json:"password"`
  PasswordExpiresAt string   `json:"password_expires_at"`
  Tags              []string `json:"tags"`
  IsAdmin           bool     `json:"is_admin"`
}
```