package storage;

import (
    "database/sql"
    "log"

    _ "github.com/lib/pq"
)



func ConnectionBD(conn string) *sql.DB {
  db, err := sql.Open("postgres", conn);
  if err != nil {
      log.Fatal("error connecting to the database: ", err);
  }
  //defer db.Close();
  return db;
}

/*
rows, err := db.Query("SELECT id, balance FROM accounts")
if err != nil {
    log.Fatal("eror3 ",err)
}
defer rows.Close()
fmt.Println("Initial balances:")
for rows.Next() {
    var id, balance int
    if err := rows.Scan(&id, &balance); err != nil {
        log.Fatal(err)
    }
    fmt.Printf("%d %d\n", id, balance)
}
*/
