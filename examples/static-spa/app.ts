const app = Golt.App();

app.static("/", "./public", true);

app.serve(3000);

