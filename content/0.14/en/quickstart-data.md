# The data template

Back to the [modding guide](index.md).

The [quick start](index.md#quick-start) made a mod from the `data` template. This page shows how its files work: what it reads, how it decides, and what it shows. It has no Lua; most mods need none.

```
.\brr mod new my-data-mod --template data
.\brr mod check mods\my-data-mod
```

| File | What it is |
|---|---|
| `manifest.json` | the mod's id, name and version, its "game" (the program `notepad.exe`), and the outputs it drives |
| `sensors.json` | the sensor `state`, which reads `state.json` |
| `rules/state.rules.json` | a tap when the alarm goes on, a tap on the left when it goes off, a flash when the door opens, a swell that follows `Level` |
| `lights.json` | the lightbar and the triggers in each situation, and the door flash |
| `feels/tap.feel.json`, `feels/swell.feel.json` | the two feels |
| `settings.schema.json` | the mod's page: effect levels, colours and triggers |
| `tests/demo.replay.jsonl` | made-up changes of `state.json`, to test the mod without Notepad |

## What it reads

`sensors.json` declares one sensor. It reads `state.json` ten times a second and names the bits of `Flags`:

```jsonc
"state": {
  "type": "file.json",
  "path": "{ModData}/state.json",
  "poll_ms": 100,
  "first_change": "silent",
  "fields": {
    "Flags": { "type": "bits", "bits": { "Alarm": 0, "Door": 1 } },
    "Level": { "type": "number" }
  }
}
```

`{ModData}` is the mod's own data folder. Each time the file changes, the sensor delivers the input `state:change`, with the old and the new values. `"first_change": "silent"` makes the first read set the values without a change, so starting Notepad plays nothing. A real game's status file works the same way. More: [Sensors](sensors.md#filejson).

## How it decides

`rules/state.rules.json` has three rules and a layer:

```jsonc
"rules": {
  "alarm_on":  { "on": "state:change", "if": "rose('Alarm')", "play": "tap" },
  "alarm_off": { "on": "state:change", "if": "fell('Alarm')", "play": { "feel": "tap", "side": "left" } },
  "door":      { "on": "state:change", "if": "rose('Door')", "do": [ { "overlay": "door_flash" } ] }
},
"layers": {
  "swell": { "feel": "swell", "if": "st and st.Level > 0", "level": "kit.clamp(st.Level, 0, 1)" }
}
```

A rule waits for its input (`on`), checks its condition (`if`, a small Lua expression), and plays a feel or runs its actions (`do`). `rose('Alarm')` holds when the Alarm bit went from off to on. A layer plays every tick while its condition holds, at the level its `level` gives. `st` is a short name for the sensor's values, set in the manifest: `"rule_env": { "st": "sensors.state" }`. More: [Rules](rules.md).

## What it shows

`lights.json` says what the lightbar and the triggers show:

```jsonc
"lightbar": {
  "calm":  { "priority": 0, "color": "calm", "bright": 0.4 },
  "alarm": { "priority": 10, "if": "st and st.flags.Alarm", "color": "alarm", "blink_hz": 2 }
},
"triggers": {
  "door": { "priority": 10, "if": "st and st.flags.Door", "l2": "soft", "r2": "firm" }
},
"overlays": {
  "door_flash": { "layer": "lightbar", "color": "flash", "ms": 600, "priority": 10 }
}
```

For each output, the row with the highest priority whose `if` holds wins. `calm`, `alarm` and `flash` are keys of the mod's `colors` setting, and `soft` and `firm` of its `triggers` setting, so the player can change them on the mod's page. More: [Situation tables](rules.md#situation-tables).

## Ideas

- Make the door tap too: add `{ "play": "tap" }` to the `do` list of the `door` rule.
- Add a lightbar row for the door: `"door": { "priority": 20, "if": "st and st.flags.Door", "color": "flash" }`.
- Add a field to `state.json` and to the sensor's `fields`, and use it in a condition.

After each change, run `mod check`, and `mod replay` with `--update` when the change is one you wanted, see [Testing](testing.md).

## On Bluetooth

Notepad never reads a controller, so the manifest says `"virtual_pad": "shared"`: Bururu changes nothing about the controller for it. On Bluetooth with the drivers the DualSense stays as it is with no game; without them Bururu opens a DualSense on Bluetooth while Notepad runs. More: [The manifest](sharing.md#the-manifest).
