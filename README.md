# Saves TIC events to TimescaleDB

## Testing

```sh
podman-compose up -d
go run cli/main.go process
declare -a fields=(IINST IINST1 IINST2 IINST3 PAPP BASE HCHP HCHC)
while sleep 1; do
    value=$((1 + RANDOM % 100))
    field=${fields[1 + $((RANDOM % ${#fields[@]}))]}
    echo "{\"ts\":$EPOCHSECONDS,\"val\":\"$(printf %03d $value)\"}" | pub -broker mqtt://localhost:1883 -topic esp-tic/status/tic/$field -username dev -password secret -qos 1
done
```
