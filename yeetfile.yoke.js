$`mkdir -p ./bin`;

yeet.setenv("GOOS", "wasip1");
yeet.setenv("GOARCH", "wasm");
yeet.run(
  "go",
  "build",
  "-o",
  "./bin/valkey-leader-yoke-airway.wasm",
  "./yoke/v1/crd/airway/",
);
yeet.run(
  "go",
  "build",
  "-o",
  "./bin/valkey-leader-yoke-flight.wasm",
  "./yoke/v1/valkey/flight",
);
