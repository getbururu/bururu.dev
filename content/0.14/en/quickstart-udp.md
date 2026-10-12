# A racing game over UDP

Back to the [modding guide](index.md).

Many racing games send telemetry while you drive: small UDP packets, many times a second, with the speed, the gear, the pedals and what each wheel touches. The `udp` template reads them and plays a clunk on each gear change, ridges while a wheel is on a kerb and a thump on a collision, and puts resistance on L2 for braking.

The template reads Racer, a made-up game, and ships a recording of its packets, so you can try everything but the last step without a game. For your game, you change the layout to its packets.

## Make it

Give `mod new` the game's name, its program (Task Manager's **Details** tab shows it while the game runs) and the port it sends to:

```
.\brr mod new my-racer --template udp --game "Gravel Rally" --exe GravelRally.exe --port 20777
.\brr mod check mods\my-racer
.\brr mod replay mods\my-racer mods\my-racer\tests\lap.replay.jsonl
```

The replay answers `ok`: the mod does what `tests\lap.expect.txt` says.

| File | What it is |
|---|---|
| `manifest.json` | the game's name and program, and the permission to listen on UDP port 20777 |
| `sensors.json` | the sensor `telemetry`: listens on `127.0.0.1:20777` while the game runs, copies every packet to `127.0.0.1:20778` for another tool, and decodes it with the layout |
| `layouts/racer.layout.json` | where each value sits in the packets |
| `rules/race.rules.json`, `rules/layers.rules.json` | the gear change, the collision, and the kerb layer |
| `lights.json` | L2 while racing |
| `feels/` | `gear_clunk`, `kerb` and `bump` |
| `settings.schema.json` | the port, the effect levels and the brake trigger on the mod's page |

`--port` sets the port everywhere, and the copy target to the next port. The manifest says `"native_dualsense": true`, for a game that drives a DualSense itself; add `--native-dualsense=false` for one that does not.

## How it reads the packets

The layout is a map of the packets. Every Racer packet starts with a 6-byte header that says which packet it is:

```jsonc
"header": { "size": 6, "fields": {
  "format":    { "at": 0, "type": "u16" },
  "packet_id": { "at": 2, "type": "u8" } } },
```

A variant is one kind of packet. The car packet holds the player's car:

```jsonc
"car": {
  "when": { "format": 1, "packet_id": 1 }, "min_len": 26,
  "fields": {
    "speed_kmh": { "at": 7,  "type": "u16", "unit": "km/h" },
    "gear":      { "at": 17, "type": "i8" },
    "surface":   { "at": 22, "type": "u8[4]", "order": ["FL", "FR", "RL", "RR"], "enum": "surface" } } },
```

Channels are values worked out from the fields, under names you choose. Rules read them as `ch['racing.speed']`:

```jsonc
"racing.speed":   { "from": "car.speed_kmh", "scale": 0.2777777777777778 },
"racing.gear":    { "from": "car.gear" },
"racing.on_kerb": { "from": "car.surface", "eq": "kerb" },
```

Edges turn a change into an input, checked on every packet:

```jsonc
"gear_change": { "on": "racing.gear", "kind": "change" },
"kerb_enter":  { "on": "racing.on_kerb", "kind": "rise", "any": true }
```

So the sensor delivers `telemetry:gear_change` on each gear change, `telemetry:event` for each event packet, and `telemetry:change` once a tick with the newest values. The rules use them:

```jsonc
"gear":      { "on": "telemetry:gear_change", "play": "gear_clunk" },
"collision": { "on": "telemetry:event", "if": "ev.tag == 'COLL'", "play": "bump" }
```

and the kerb layer and the L2 effect read the newest car packet as `tel.car` (the manifest's `rule_env` names `tel` for `sensors.telemetry`): `tel.car.surface` and `tel.car.speed_kmh`. The whole layout language: [Binary layouts](layouts.md).

## Change it

Make the gear clunk deeper. In `mods\my-racer\feels\gear_clunk.feel.json`, change `"f0_hz": 190` to `"f0_hz": 160`.

```
.\brr mod check mods\my-racer
```

The replay now says it plays differently: the clunk sounds different. Write it:

```
.\brr mod replay mods\my-racer mods\my-racer\tests\lap.replay.jsonl --update
.\brr mod replay mods\my-racer mods\my-racer\tests\lap.replay.jsonl
```

More ideas:

- A thump at the rev limiter: the layout already has the edge `rev_limit`, so add `"rev": { "on": "telemetry:rev_limit", "play": "bump" }` to `rules/race.rules.json`.
- More values: find their offsets in your game's published UDP format, add them as fields, then as channels or in rules.

## In the game

1. Start Bururu. The **Mods** page lists your mod as **Needs your consent**: a mod you made asks before it listens on a port. **Turn on** shows the port and the copy target.
2. In the game's settings, turn UDP telemetry on and send it to `127.0.0.1`, port 20777.
3. Drive. The sensor takes the port only while the game runs.

A game that drives the DualSense itself fights Bururu for the triggers and haptics on the cable. The mod's first page says to turn the game's own vibration and trigger effects off, or to pick **Direct** for the game, which leaves it the controller while it runs. With the drivers for the virtual controller there is nothing to do: the game sees Bururu's virtual controller.

To test with your own driving, record it: `.\brr mod record my-racer --out lap1.replay.jsonl` records the mod's packets until Ctrl+C, see [Testing](testing.md).

## Pack it

```
.\brr mod pack mods\my-racer --out .
```

A feel pack that works with several games needs them to publish the same channel names, see [Vocabularies](sharing.md#vocabularies).
