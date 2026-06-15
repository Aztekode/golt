(async () => {
  const db = Golt.db.connect("sqlite", "./app.db");

  await db.exec(`
    CREATE TABLE IF NOT EXISTS users (
      id INTEGER PRIMARY KEY AUTOINCREMENT,
      name TEXT NOT NULL
    );
  `);

  await db.exec(`INSERT INTO users (name) VALUES ('Ada');`);

  const users = await db.query(`SELECT id, name FROM users ORDER BY id DESC LIMIT 10;`);
  console.log(users);
})();
