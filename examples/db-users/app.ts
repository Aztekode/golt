const db = Golt.db.connect("sqlite", "./app.db");

db.exec(`
  CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL
  );
`).then(async () => {
  const existing = await db.query("SELECT COUNT(*) AS c FROM users;");
  const count = Number(existing?.[0]?.c ?? 0);
  if (count === 0) {
    await db.exec("INSERT INTO users(name) VALUES (?)", "Ada");
    await db.exec("INSERT INTO users(name) VALUES (?)", "Grace");
  }
});

const app = Golt.App();

app.use(Golt.logger({ format: "dev" }));

app.get("/users", async (ctx) => {
  const users = await db.query("SELECT id, name FROM users ORDER BY id ASC;");
  ctx.Json({ users });
});

app.serve(3000);

