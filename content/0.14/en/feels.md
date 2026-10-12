# Feels

A feel is what the haptics play: a click, a swell, a rattle. On the cable it plays on the two haptic actuators as a waveform; on Bluetooth the same waveform at a lower rate; on a controller with rumble only, on the motors. A mod plays a feel by name. Bururu mixes every feel that plays, applies the player's effect levels and **Strength**, and keeps the sum from distorting.

Each feel is one file, `feels/<name>.feel.json`. Every field is in the [file reference](files.md#feel). To design feels by ear: [Feels in REAPER](reaper.md).

## A feel file

```jsonc
{
  "feel": 1,
  "name": "tap",
  "kind": "shot",
  "voices": [
    { "id": "click", "type": "synth", "wave": "Sine", "f0_hz": 180, "f1_hz": 150,
      "attack_ms": 2, "decay_ms": 45, "decay_curve": 2, "level": 0.5 }
  ]
}
```

| Field | What it is |
|---|---|
| `name` | the name rules and Lua play; default: the file's name |
| `kind` | `shot`: plays once. `layer`: held; its level and params are set every tick |
| `voices` | the voices, added up; a layer has exactly one |
| `params` | named inputs a rule or Lua sets, see [Params and modulation](#params-and-modulation) |
| `mix` | `gain`: the effect level it follows; `duck`: a bus it ducks |
| `rumble` | what it does on the motors, see [Motors](#motors) |
| `meta` | `description`, `tags`, and `source` (see [Feels in REAPER](reaper.md)) |

## Voices

A `synth` voice is the same as the Bururu Voice in REAPER, field for field. Times are in ms. A missing field takes its default; a value out of range is an error.

| Field | Range | Default | What it does |
|---|---|---|---|
| `id` | | | the voice's name; patches name voices by it |
| `wave` | | `Sine` | `Sine`, `Square`, `Triangle`, `Saw`, `Noise`, `NormNoise`, `BandNoise`, `Grains` or `Metal` |
| `f0_hz` | 0-5000 | 170 | the frequency it starts at |
| `f1_hz` | 0-5000 | 0 | the frequency it sweeps to; 0: no sweep |
| `sweep`, `sweep_ms` | | `exp`, 0 | `exp` or `linear`, and how long; 0: the whole voice |
| `delay_ms` | 0-60000 | 0 | silence before it starts |
| `attack_ms`, `attack_curve` | 0-60000, 0.1-8 | 0, 1 | the rise; 1 is linear |
| `hold_ms` | 0-600000 | 0 | full level after the attack |
| `decay_ms`, `decay_curve` | 0-60000, 0.1-8 | 0, 1 | the fall to the sustain level |
| `sustain`, `sustain_ms` | 0-1, 0-600000 | 0, 0 | the level after the decay, and how long it holds |
| `release_ms`, `release_curve` | 0-60000, 0.1-8 | 0, 1 | the fade at the end |
| `level` | 0-2 | 0.5 | how strong it is |
| `left`, `right` | 0-1 | 1, 1 | the left and right actuator |
| `trem_hz`, `trem_depth` | 0-200, 0-1 | 0, 0 | tremolo |
| `gate_hz`, `gate_duty` | 0-200, 0-1 | 0, 0.5 | on and off at this rate, on for this share |
| `band_hz` | 1-4000 | 80 | `BandNoise`: the width of the band |
| `grain_rate`, `grain_cycles`, `grain_jitter`, `grain_spread` | | 70, 2, 30, 20 | `Grains`: grains per second, cycles per grain, jitter and spread in % |
| `seed` | | 1 | the same seed plays the same noise every time |
| `amount` | | | `ref`, `level_per`, `level_max`: the level grows with a play's `amount`, such as a hit's damage |
| `mod` | | | modulation, below |

## Params and modulation

Params let one feel change with the game: a whir that rises with speed. A feel declares its params, and its voices say what each one changes:

```jsonc
"params": { "spin": { "min": 0, "max": 1, "default": 0 } },
"voices": [
  { "id": "whir", "type": "synth", "wave": "Saw", "f0_hz": 90, "level": 0.3,
    "mod": [ { "target": "f0_hz", "source": "spin", "add": 60 },
             { "target": "level", "source": "spin", "add": 0.23 } ] }
]
```

A rule sets them with `"params": { "spin": "state.spin" }`, Lua with `feel.layer("spin", "whir", 1, { params = { spin = s.spin } })`.

| Form | What it does |
|---|---|
| `{ "target", "source", "add": k }` | target + k x source |
| `{ "target", "source", "mul": k }` | target x (k x source) |
| `{ "target", "source", "map": [[x, y], ...] }` | a drawn curve; `op` is `set`, `add` or `mul`, `curve` is `linear` or `smooth` |

- Targets: `level`, `f0_hz`, `f1_hz`, `trem_hz`, `trem_depth`, `gate_hz`, `gate_duty`, `band_hz`, `grain_rate`, `grain_spread`, `left`, `right`, and `time` (stretches a one-shot).
- Sources: a param; `level`, `scale`, `amount`, `t` (seconds since the feel started); or an LFO: `{ "lfo": "sine", "hz": 0.5 }`, with `sine`, `triangle`, `square` and `random`.
- A one-shot reads its params once, as it starts. A layer follows them every tick, and glides, so changes never click.

## Mix

- **Effect levels**: each feel follows a level the player sets on the mod's page, from the mod's setting with the role `feel_gain`. A level of 0 turns the feel off.
- **Ducking**: a one-shot with `"mix": { "duck": { "bus": "turn", "ms": 350 } }` marks the bus `turn` for 350 ms. The mod's Lua reads `t.ducked.turn` and turns a layer down meanwhile.
- At most 64 one-shot voices sound at once; a new one replaces the oldest.

## Motors

With rumble only, or the player's **Mode** set to **Rumble**, feels play on the two motors. Each feel's strengths and length come from the mod's setting with the role `feel_rumble`, so the player can change them. A feel without an entry plays as the manifest says: `"feels": { "rumble_default": "derived" }` (the default) works one out from the voices; `"none"` plays nothing.

A one-shot's `rumble` gives a pattern for the motors: `{ "pattern": "once" }`, `{ "pattern": "repeat", "count": 2, "gap_ms": 150 }`, or `{ "pattern": "sweep", "sides": ["left", "right"], "gap_ms": 300 }`.

`"rumble": { "triggers": "r2" }` (or `"l2"`, `"both"`) marks a one-shot for the triggers: on a controller with trigger motors (an Xbox One or Series controller on Windows, `t.caps.impulse`), each of its motor pulses also pushes that trigger, at the pulse's stronger motor level and for 120 ms at most, scaled by the player's trigger motor strength. Played on the left, only L2 pushes; on the right, only R2. A layer's mark is ignored: it would hold the motor steady. Keep these for moments the finger should feel, such as a shot or a gear change.

## On Bluetooth

On Bluetooth the actuators get the waveform at 3000 samples a second, in 8 bits. So nothing above about 1 kHz arrives, quiet tails turn into grit or silence, and peaks above 0.8 are squashed. Feels between about 100 Hz and a few hundred Hz, with clear starts, carry best on both connections. A mod can read `t.caps.voice_rate` (48000 on the cable, 3000 on Bluetooth). On macOS a DualSense on its cable has no wave haptics (`t.caps.voices` is false, `voice_rate` 0), so a mod's one-shots play their `rumble` patterns on its motors there.

## The taste lint

```
.\brr feel check mods\my-data-mod --taste
```

It renders each feel as the cable and Bluetooth play it, and warns about what tends to feel bad: a steady tone over 1.5 s (`steady`), energy below 100 Hz, where the actuators rattle (`low`), a feel Bluetooth turns silent (`bt_quiet`), energy above 1 kHz (`high`), peaks above 0.8 (`peak`). They are warnings: break them on purpose when you want to. Short clicks and swells that come and go feel better over a long session than steady tones.

## Hearing a feel

```
.\brr feel render mods\my-data-mod\feels\tap.feel.json --out tap.wav
```

This writes the feel as a WAV file, the two actuators as left and right, at 48 kHz (`--profile dualsense-bt` as Bluetooth plays it). Played through the DualSense's audio device on the cable (channels 3 and 4), the file feels as the feel does.

## Changing another mod's feels

A feel's full name is `<mod>/<name>`. An add-on replaces one with `overrides/<mod>/feels/<name>.feel.json`, or patches it with `overrides/<mod>/feels/<name>.feel.patch.json`:

```jsonc
{
  "voices": {
    "click": { "level": 0.2 },
    "sparkle": { "type": "synth", "wave": "Sine", "f0_hz": 320, "decay_ms": 30, "level": 0.1 }
  }
}
```

A patch changes the fields it names, and `null` removes. Voices are named by id: a named voice changes field by field, `null` removes it, a new id adds one. A replacement must declare every param the original has. See [Add-ons](sharing.md#add-ons).
