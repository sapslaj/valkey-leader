$`mkdir -p ./bin`;

yeet.setenv("CGO_ENABLED", "0");
go.build("-o", "./bin/valkey-leader");
