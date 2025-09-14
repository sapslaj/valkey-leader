#!/usr/bin/env -S just --justfile

yeet:
  go tool github.com/TecharoHQ/yeet/cmd/yeet

yoke: yoke-flight yoke-airway

yoke-airway:
  mkdir -p ./bin/
  GOOS=wasip1 GOARCH=wasm go build -o ./bin/valkey-leader-yoke-airway.wasm ./yoke/v1/crd/airway/

yoke-flight:
  mkdir -p ./bin/
  GOOS=wasip1 GOARCH=wasm go build -o ./bin/valkey-leader-yoke-flight.wasm ./yoke/v1/valkey/flight/
