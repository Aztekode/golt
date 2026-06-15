const app = Golt.App();

app.use(Golt.logger({ format: "dev" }));

app.get("/", (ctx) => {
  ctx.Json({
    message: "Hello from Golt examples!",
    example: "hello-http",
  });
});

app.serve(3000);

