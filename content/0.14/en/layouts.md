# Binary layouts

Back to the [modding guide](index.md).

A racing game's telemetry comes as UDP packets: rows of bytes, each value at a known offset. A layout file (`layouts/<name>.layout.json`) says which bytes are which value, so a `udp.listen` sensor can decode them. Rules and Lua only see the numbers.

Declare only the fields you use. The game's published UDP format gives the offsets and types. The [udp template](quickstart-udp.md) has a whole layout. Every field is in the [file reference](files.md#layout).

## The shape of a layout

```jsonc
{
  "layout": 1,
  "endian": "little",
  "header":   { "size": 6, "fields": { ... } },
  "variants": { "<name>": { ... } },
  "enums":    { "surface": { "0": "tarmac", "1": "kerb", "*": "other" } },
  "channels": { "racing.speed": { ... } },
  "edges":    { "gear_change": { ... } },
  "gate":     { "variant": "car", "field": "in_race", "ne": 0 },
  "unknown":  "report"
}
```

- `endian` is the byte order of every field (`little` or `big`).
- `header` is the part every packet starts with; `size` is its length.
- `variants` are the kinds of packets, each with when it applies and its fields.
- `enums` name values, `channels` give fields shared names, `edges` turn changes into inputs, and `gate` drops data while the game is in a menu.

## Fields

```jsonc
"speed_kmh": { "at": 7,  "type": "u16", "unit": "km/h" },
"throttle":  { "at": 9,  "type": "f32" },
"gear":      { "at": 17, "type": "i8" },
"surface":   { "at": 22, "type": "u8[4]", "order": ["FL", "FR", "RL", "RR"], "enum": "surface" }
```

`at` is the offset in bytes from the packet's start.

| Type | What it is |
|---|---|
| `u8`, `i8`, `u16`, `i16`, `u32`, `i32`, `u64`, `i64` | whole numbers of 1, 2, 4 and 8 bytes, unsigned and signed |
| `f32`, `f64` | floating-point numbers; NaN and infinity read as 0 |
| `bool8`, `bool32` | true when not 0 |
| `char[N]`, `wchar[N]` | text of N bytes, or N UTF-16 characters |
| `T[N]`, `T[N][M]` | a list of N values, or N lists of M |
| `bits32` | a number with named bits: `"bits": { "abs": 0, "tc": 1 }` |

| Option | What it does |
|---|---|
| `endian` | this field's byte order |
| `scale`, `offset` | the value times `scale`, plus `offset` |
| `clamp` | `[low, high]` |
| `enum` | names its values from an `enums` map; unnamed values read as `"*"`, or as the number |
| `order` | a four-wheel list's order in the game; it is turned into front left, front right, rear left, rear right |
| `redact` | `true`: recordings hold zeros in its place, for other players' names |

## Variants

Variants are tried in the order written; the first whose `when` holds takes the packet.

| `when` key | Holds when |
|---|---|
| a header field | it has this value, or one of a list: `"packet_id": 1`, `"format": [1, 2]` |
| `length` | the packet has this length, or one of a list |
| `magic` | the packet holds these bytes at an offset: `{ "at": 0, "text": "RACE" }` |

`min_len` is the shortest packet the variant takes. Variants let one layout read several versions of a game, or several games.

### Arrays

Many games send every car in one packet. `array` reads one element of it:

```jsonc
"cars": {
  "when": { "format": 1, "packet_id": 3 },
  "array": { "at": 6, "count": 20, "stride": 40, "pick": "header.player_car" },
  "fields": { "speed_kmh": { "at": 0, "type": "u16" } }
}
```

`at` is where the first element starts, `stride` the bytes per element, `count` how many (a number or `"header.<field>"`), and `pick` a header field that says which element is the player's. Offsets inside count from the element's start. Without `pick`, each field becomes a list over every element.

### Event packets

A packet whose meaning depends on a tag inside it uses `tag` and `cases`, and `emit` makes each packet an input:

```jsonc
"event": {
  "when": { "format": 1, "packet_id": 2 },
  "tag": { "at": 6, "type": "char[4]" },
  "cases": { "COLL": { "fields": { "other_car": { "at": 10, "type": "u8" } } }, "LAPS": {}, "PITS": {} },
  "emit": "event"
}
```

Each such packet is the input `<sensor>:event` with `tag` and its case's fields; a rule tests `ev.tag == 'COLL'`. With `cases`, only the listed tags come.

## Channels

Channels are values worked out from the fields, under names you choose, written `<group>.<name>`. When they use the names of a [vocabulary](sharing.md#vocabularies), rules and feel packs work for every game that uses them:

```jsonc
"channels": {
  "racing.speed":   { "from": "car.speed_kmh", "scale": 0.2777777777777778 },
  "racing.on_kerb": { "from": "car.surface", "eq": "kerb" },
  "racing.in_race": { "from": "car.in_race", "ne": 0 }
}
```

A channel is one entry, or a list tried in order, where the first whose packets came recently gives the value. An entry takes `from` (`<variant>.<field>`), or `hypot` (the length of several fields), then in order `div`, `abs`, `scale` and `offset`, `clamp`, `eq` or `ne`, and `any` (over a four-wheel list: true when any wheel holds). Rules read four-wheel channels as `ch['racing.on_kerb'][1]` for the front left.

## Edges

An edge turns a change into an input, checked on every packet:

```jsonc
"edges": {
  "gear_change": { "on": "racing.gear", "kind": "change" },
  "kerb_enter":  { "on": "racing.on_kerb", "kind": "rise", "any": true },
  "rev_limit":   { "on": "car.rpm", "kind": "threshold", "above": 11800, "below": 11500 }
}
```

| `kind` | Fires when |
|---|---|
| `change` | the value changes |
| `rise` | it turns true, or from 0 to another number |
| `fall` | it turns false, or 0 |
| `threshold` | it goes above `above`; again only after it went below `below` |

On a four-wheel value, `"any": true` makes it one value; `rise`, `fall` and `threshold` need it. The first packet after a start or a pause never fires. The input is `<sensor>:<edge>`, with `value` and `old`.

## Gate and unknown packets

While the `gate`'s field fails its test (`eq` or `ne`), packets count for nothing, so a game in a menu plays no kerbs. With `"unknown": "report"` (the default), the sensor says "unknown version" when packets come and none has matched a variant for 2 seconds.

## Testing a layout

`mod check` reads the layout and finds offsets that do not fit, unknown types and names. Record packets while you drive (`mod record`) and replay them at your desk (`mod replay`), see [Testing](testing.md). A short or broken packet is counted and dropped.
