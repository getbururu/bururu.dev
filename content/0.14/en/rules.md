# Rules

Back to the [modding guide](index.md).

Most of a mod is data: rules that play a feel when something happens, layers that play while something holds, and tables for the lights and triggers. Conditions are small Lua expressions. Every field is in the [file reference](files.md#rules).

## Rules files

`rules/*.rules.json` hold rules and layers. A mod may have as many files as it likes:

```jsonc
{
  "format": 1,
  "defaults": { "if": "st and not st.flags.Paused" },
  "rules":  { "<id>": { ... } },
  "layers": { "<id>": { ... } }
}
```

`defaults.if` is a condition every rule and layer of the file must also meet. Ids may not hold `/` or spaces. Add-ons and the player name a rule as `<mod>/<id>`.

## Rules

A rule waits for an input, checks a condition and plays a feel or runs its actions:

```jsonc
"alarm_on":  { "on": "state:change", "if": "rose('Alarm')", "play": "tap" },
"collision": { "on": "telemetry:event", "if": "ev.tag == 'COLL'", "play": "bump" }
```

| Field | What it does |
|---|---|
| `on` | the input it waits for, see [Selectors](#selectors) |
| `if` | the condition |
| `play` | a feel to play, see [Playing a feel](#playing-a-feel) |
| `do` | actions, in order, see [Actions](#actions) |
| `order` | its place among the rules of one input (default 1000) |
| `group`, `priority` | of the rules of one group that match, only the highest priority fires, see [Groups](#groups-priority-and-stop) |
| `cooldown_ms` | after it fired, it waits this long before it fires again |
| `stop` | when it fires, no later rule sees the input |
| `replay` | it also fires while a file sensor catches up on what was written before Bururu started; such a rule should only set state |

### Selectors

An input is named `<sensor>:<name>`: the sensor's id in `sensors.json`, then what happened.

| `on` | Matches |
|---|---|
| `events:lap_done` | one input of one sensor |
| `events:*` | every input of the sensor |
| `state:change` | a change of a sensor's values, with the old and new values |
| `telemetry:gear_change` | an edge the sensor declares |
| `binds:action=horn` | the input `binds:action` whose field `action` is `horn` |
| `game:started`, `game:closed`, `game:check` | the game starts, ends, or is looked for (every 3 s) |
| `sensor:status` | a sensor's state changed (fields `id`, `state`, `text`) |

`mod check` checks that the sensor exists and, where it can, that the input does. For one input, the mod's Lua runs first, then the matching rules.

### Playing a feel

`play` is a feel's name, or how it plays:

```jsonc
"play": "tap"
"play": { "feel": "tap", "side": "left" }
"play": { "feel": "hit", "pan": "ev.pan", "scale": "ev.strength", "amount": "ev.damage" }
```

| Field | What it does |
|---|---|
| `feel` | the feel: its name, or `<mod>/<name>` for another mod's |
| `side` | `both`, `left` or `right` |
| `scale` | the strength; 1 is as made, 0 or less plays nothing |
| `pan` | from -1 (left) to 1 (right) |
| `amount` | what the feel's amount-scaled voices scale by, such as the damage of a hit |
| `delay_ms` | start this much later |
| `gain` | the effect level it follows (default: the feel's own) |
| `rumble` | the motor setting it uses (default: the feel's own) |
| `params` | values for the feel's params, see [Params](feels.md#params-and-modulation) |

A number field takes a number or an expression in a string (`"scale": "ev.strength"`). A text field takes text, or an expression as `{"expr": "..."}`. A feel chosen by an expression lists the feels it may choose, so each is checked when the mod loads:

```jsonc
"play": { "feel": { "expr": "ev.big and 'thump' or 'tap'", "one_of": ["thump", "tap"] } }
```

A feel name is looked for in the rule's own mod, then in the game mod it changes, then in its libraries. A name found nowhere is an error.

### Actions

| Action | What it does |
|---|---|
| `{ "play": ... }` | plays a feel |
| `{ "overlay": "door_flash" }` | starts a timed overlay, see [Overlays](#overlays) |
| `{ "overlay_stop": "door_flash" }` | stops it |
| `{ "set": { "state.alarm": "true" } }` | sets a value in the game mod's `bururu.state` |
| `{ "channel": { "racing.boost": "ev.Boost / 100" } }` | publishes a vocabulary channel, see [Vocabularies](sharing.md#vocabularies) |
| `{ "log": "Gear {ev.value}" }` | writes a log line; each `{...}` is an expression |
| `{ "notice": "Telemetry is off" }` | shows a message in the window |

A value that cannot be worked out skips that action only.

### Groups, priority and stop

Rules that match one input run by `order`, then in the order of the files and as written. Each rule whose condition holds fires, unless:

- **It is in a group**: of the rules of one `group` that match and hold, only the one with the highest `priority` fires. A tie goes to the mod that loads later. So an add-on can outrank a game mod's rule under a condition of its own.
- **It is cooling down** after its last fire.
- **An earlier rule stopped the input** with `"stop": true`.

## Layers and stages

A layer is a continuous feel: every tick, each layer whose condition holds plays at its level. A layer that stops holding fades out.

```jsonc
"layers": {
  "swell": { "feel": "swell", "if": "st and st.Level > 0", "level": "kit.clamp(st.Level, 0, 1)" }
}
```

| Field | What it does |
|---|---|
| `feel` | a `layer` feel |
| `if` | the condition |
| `level` | 0 to 1 (default 1) |
| `params` | the feel's params, set every tick |
| `side` | `both`, `left` or `right` |
| `key` | its name in the mix (default: its id); two layers with one key: the last one plays |
| `gain` | the effect level it follows (default: its key) |
| `rumble` | the motor setting, or `false` for nothing on the motors |
| `stage`, `order` | when it is worked out, see below |

Layers are worked out in stages, each a name with an `order` (`"stages": { "engine": { "order": 10 } }` in a rules file); a layer without one is in `default`. A mod with Lua can work a stage out at a moment of its own with `rules.stage("engine")`. A mod without Lua never needs stages.

## Situation tables

`lights.json` says what the lightbar, the triggers, the player LEDs and the mic LED show in each situation:

```jsonc
{
  "format": 1,
  "lightbar": {
    "calm":  { "priority": 0, "color": "calm", "bright": 0.4 },
    "alarm": { "priority": 10, "if": "st and st.flags.Alarm", "color": "alarm", "blink_hz": 2 }
  },
  "triggers": {
    "door": { "priority": 10, "if": "st and st.flags.Door", "l2": "soft", "r2": "firm" }
  },
  "leds": {
    "level": { "priority": 0, "count": "gomath.ceil((st and st.Level or 0) * 5)" }
  },
  "mic": {
    "alarm": { "priority": 0, "if": "st and st.flags.Alarm", "mic": "pulse" }
  },
  "overlays": {
    "door_flash": { "layer": "lightbar", "color": "flash", "ms": 600, "priority": 10 }
  }
}
```

Each table needs its output in the manifest's `outputs` (`lightbar`, `triggers`, `player_leds`, `mic`). Each field is decided on its own: the row with the highest `priority` whose `if` holds and that sets the field wins it. So a row may set only the brightness and leave the colour to a lower row. A row without `if` always holds.

| Table | Row fields |
|---|---|
| `lightbar` | `color`: a key of the mod's colour setting; `rgb`: an expression giving `{r, g, b}`; `bright`: a share of the player's brightness; `bright_abs`: 0 to 255; `blink_hz`; `off` |
| `triggers` | `l2`, `r2`, `both`: a key of the mod's trigger setting, or `none`; `l2_raw`, `r2_raw`: an effect, such as `{"mode": "FEEDBACK", "params": [3, 3]}` |
| `leds` | `count`: 0 to 5 |
| `mic` | `mic`: `on`, `pulse` or `off` |
| any | `final`: while the row holds, no overlay shows |

Colour and trigger keys name entries of the mod's settings with the roles `color_table` and `trigger_table`, so the player can change them on the mod's page.

### Trigger effects

A trigger effect, in `l2_raw` and `r2_raw` or in an entry of the mod's trigger setting, is a mode with its values, for example `{"mode": "WEAPON", "params": [2, 5, 6]}`:

| Mode | Values | In the window |
|---|---|---|
| `OFF` | none | **Free** |
| `FEEDBACK` | start 1-9, strength 1-8 | **Wall** |
| `WEAPON` | start 2-7, end 3-8, strength 1-8 | **Click** |
| `VIBRATION` | start 1-9, amplitude 1-8, frequency 1-40 | **Rattle** |
| `SLOPE_FEEDBACK` | start 1-8, end 2-9, start strength 1-8, end strength 1-8 | **Ramp** |
| `MULTIPLE_POSITION_FEEDBACK` | 10 values, 0-8 | **Wall by position** |
| `MULTIPLE_POSITION_VIBRATION` | frequency 1-40, then 10 values, 0-8 | **Rattle by position** |

Positions count from 0 (at rest) to 9 (pressed all the way). In `lights.json`, an unknown mode, the wrong number of values or a value out of its range is an error.

### Overlays

An overlay is a timed flash on top of the tables: the lightbar in a colour, or both triggers in an effect.

| Field | What it does |
|---|---|
| `layer` | `lightbar` or `triggers` |
| `color` | lightbar: a colour key |
| `trigger` | triggers: a trigger key, for both triggers |
| `ms` | how long it shows |
| `blink_hz` | lightbar: blink at this rate |
| `priority` | of the overlays that show, the highest wins |
| `if` | it shows only while this holds |

A rule starts one with `{ "overlay": "<id>" }`, Lua with `out.overlay("<id>")`.

## Conditions and values

Conditions (`if`) and values (`level`, `scale`, `count`, `{"expr": ...}`) are Lua expressions:

```
rose('Alarm')
st and st.flags.Door
gomath.min(1, 0.4 + (ch['racing.speed'] or 0) / 60)
```

An expression may use names, fields, `and`, `or`, `not`, comparisons, arithmetic, `..`, `#`, and calls to `lib.*`, `kit.*`, `gomath.*`, `bit.*`, `changed`, `rose` and `fell`. It may not define functions or assign. Anything more belongs in Lua: a rule can call the mod's own Lua module as `lib.<name>(...)`.

| Name | What it is |
|---|---|
| `ev` | the input's fields, plus `at` (ns since the session started) and `live` |
| `old`, `new` | the values before and after a change |
| `changed('X')`, `rose('X')`, `fell('X')` | the field or bit `X` changed; went from off to on; from on to off |
| `sensors.<id>` | a sensor's latest values |
| `st`, ... | short names the manifest's `rule_env` gives: `"st": "sensors.state"` |
| `lib` | the Lua module the manifest's `rule_env` names: `"lib": "lua/lib.lua"` |
| `state` | the game mod's `bururu.state` |
| `settings` | this mod's settings |
| `t` | the tick: `t.now` (ns), `t.unix_ms` (the clock, for blinks), `t.dt` |
| `ch` | vocabulary channels: `ch['racing.rpm']` |
| `kit`, `gomath`, `bit` | helpers, see the [Lua API](lua-api.md#the-kit-table) |

A sensor's named bits are read through a twin of their field with a lower-case first letter: `Flags` is the number, `flags.Alarm` one bit.

Names are checked when the mod loads. An error names its file, line and column, and says "did you mean" when a name is close.

### When a value is missing

Game data is often missing: a field the game did not write, a sensor that has not read yet. An expression that fails (a `nil` compared with a number) reads as false, and a value that fails skips its action. The rule stays on. Keep conditions safe with `st and ...`, `(x or 0)` and `kit.has(x)`.

## The player's overrides

A player can switch a rule or layer off, or swap its feel, in the mod's values file `mod-settings\<id>.json`:

```jsonc
{ "_rules": { "my-data-mod/alarm_off": { "enabled": false },
              "my-data-mod/alarm_on":  { "play": "swell" } } }
```

A swapped feel replaces every feel the rule plays. An override for a rule that does not exist is a warning.

## When to use Lua

Use rules and tables while you can: they are short, an add-on can patch each by id, and a typo is caught at load. Use Lua when a moment depends on what came before (a timer, a count, a value that builds up), when a value must change only at certain moments, or when a few lines of logic would need many rules. A mod can have both: rules read the Lua's `bururu.state` as `state` and call its module as `lib`. See [A mod in Lua](quickstart-lua.md).

## The phases of a tick

Every 25 ms Bururu runs one tick. A mod's Lua and rules run at fixed places in it:

| Phase | What runs |
|---|---|
| `init` | each mod's Lua entry, once, when the mod loads |
| `read` | for each input from the game's files and packets: the Lua, then the rules; then the game mod's `facts` |
| `before_screen`, `screen` | what the screen sensor should read; then its inputs |
| `idle` | `on_idle`, while the game is not active or haptics are off |
| `drive` | `on_pad`, `on_tick` and the layers, while the game is active |
| `frame` | `before_frame`, the situation tables, the overlays, `on_frame`; then Bururu sends what changed to the controller |

Active means the game runs, the game mod says the player is in control, and the effects are not paused. While not active, feels stop and the controller rests.
