# Sensors

A sensor is Bururu code that reads something for a mod: a game's files, the screen, its control preset, network packets, the game's own sound, what it sends to its controller, and for some games their memory. A mod reads nothing itself. It declares the sensors it needs in `sensors.json`, and gets what they read in two ways:

- An **input** is one thing that happened, delivered once: `state:change`, `telemetry:gear_change`. Its name is the sensor's id, a colon, and what happened. Rules wait for inputs with `on`, Lua with `bururu.on`.
- A **snapshot** is a sensor's latest values, read any time as `sensors.<id>`, or by a short name the manifest's `rule_env` gives it.

## Which sensor to use

Pick what the game gives most reliably, and use more than one when each gives a different part:

- **A single-player game that the mod list marks offline with no anti-cheat:** its [memory](#memory) first. It holds the game's state as numbers, such as health, stamina and speed. Then the game's own [sound](#sound), which tells exactly when a hit, a parry or an explosion happens, and the [screen](#screen) for what the HUD shows. Log files last.
- **An online game, or a single-player game with an anti-cheat:** its log files or telemetry packets first ([file.text](#filetext), [file.jsonl](#filejsonl), [udp.listen](#udplisten)), when it writes them. Then its [sound](#sound), then the [screen](#screen). Bururu never reads the memory of these games.
- **A game that rumbles the controller:** its rumble, from [game.output](#gameoutput), follows every moment the game's makers gave a rumble. A mod can play its own feels in place of that rumble, and add its own for the moments other sensors find.
- **No mod at all:** Bururu's own sound haptics follow the game's sound by themselves, with no mod. A mod is how a game gets designed feels.

State and moments often come from different sensors: a health bar from the memory or the screen, the hit that took it from the sound. The sound sensor and the screen do not depend on an anti-cheat; the sound sensor does not depend on the screen's size or the HUD's layout either.

## Declaring a sensor

```jsonc
{
  "sensors": 1,
  "items": {
    "state": {
      "type": "file.json",
      "path": "{ModData}/state.json",
      "poll_ms": 100,
      "fields": { "Flags": { "type": "bits", "bits": { "Alarm": 0, "Door": 1 } } }
    }
  }
}
```

Each item is keyed by the sensor's id. Fields every kind has:

| Field | What it does |
|---|---|
| `type` | the kind, below |
| `order` | its place among the sensors that read at the same time: bindings first, then the game's files and packets, then the screen |
| `poll_ms` | for file sensors: how often to read; 0 is every tick (25 ms) |
| `when` | `game.running` or `window.in_front`: the sensor runs only then |
| `stale_ms` | values older than this read as missing |
| `redact` | fields left out of recordings, by input name, see [Recording a session](testing.md#recording-a-session) |
| `patchable` | the fields an add-on may change |

Paths use tokens, so one mod works on every PC:

| Token | What it is |
|---|---|
| `{SavedGames}`, `{Documents}`, `{LocalAppData}`, `{RoamingAppData}`, `{Home}` | your Windows folders |
| `{SteamApp:<app id>}` | where Steam installed that game |
| `{GameDir}`, `{GameDir:<game id>}` | the game's install folder, as Bururu found it on this PC, see [Finding the game](sharing.md#finding-the-game) |
| `{ModData}` | the mod's own data folder, `mod-data\<id>\` |
| `{ModDir}` | the mod's own files |
| `{path:<name>}` | a path the manifest declares under `paths`; the player can point it elsewhere |
| `{setting:<path>}` | a value of the mod's settings |

A mod must name a folder below a Windows folder, never the folder itself, and the player agrees to each folder before the mod runs, see [Permissions](sharing.md#permissions). A path that does not exist on this PC makes its sensor "unavailable here"; the mod still loads.

On Linux, for a game that runs through Proton or Wine, the Windows folders are the ones inside its prefix, see [Games on Linux](sharing.md#games-on-linux). On macOS, for a Windows game in a CrossOver bottle, they are the ones inside the bottle, see [Games on macOS](sharing.md#games-on-macos).

Every field of a declaration:

| Field | Type | What it is |
|---|---|---|
| `bind` | string | udp.listen: the address to bind. |
| `catch_up` | one of state_only, skip, play | file.jsonl, file.text: what the first file's lines do. |
| `debug_dir` | string | screen: where debug captures go. |
| `dir` | string | file.jsonl, file.text, bindings: the folder, with tokens. |
| `edges` | object of object by key | file.json: inputs when a value moves, by name. |
| `edges.*.above` | number | threshold: fires as the value passes above this. |
| `edges.*.below` | number | threshold: armed again below this; default: above. |
| `edges.*.kind` | one of change, rise, fall, threshold | change: any change; rise: from 0 or false to anything else; fall: back to 0 or false; threshold: it passes above. |
| `edges.*.on` | string | A field, or one of its named bits: Flags.Docked. |
| `fallback_file` | string | bindings: the file to try next, with {preset}. |
| `fields` | object of object by key | file.json: typed fields and named bits. |
| `fields.*.bits` | object of integer by key | bits: named bits, name -> bit index. |
| `fields.*.type` | one of bits, int, int?, int[], number, number?, number[], string, string?, string[], bool, bool?, object, any | Its type; with ? a missing field reads as nil, with [] it is a list. |
| `file` | string | bindings: the preset file, with {preset}, {major} and {minor}. |
| `first_change` | one of silent, change | file.json: the first good read sets the snapshot without a change. |
| `forward` | list of string | udp.listen: forward each datagram to these targets. |
| `glob` | string | file.jsonl, file.text: the files to follow. |
| `layout` | string | udp.listen, bindings, screen: the layout file in the mod. |
| `lines` | object of object by key | file.text: line rules with Go regular expressions. |
| `lines.*.re` | string | A Go regular expression with named captures. |
| `name_field` | string | file.jsonl: the field that names the event. |
| `order` | integer | Its place among the sensors that read at the same time. |
| `palette` | object | screen: the palette files. |
| `palette.learned_file` | string | The learned palette file. |
| `palette.matrix_file` | string | The colour matrix file. |
| `patchable` | list of string | Parameters other mods may patch. |
| `path` | string | file.json: the file, with tokens. |
| `pick` | one of newest_mtime | file.jsonl, file.text: which file. |
| `pointer` | object | bindings: where the preset name is. |
| `pointer.fallback_line` | integer | The line to try next. |
| `pointer.files` | list of string | Files to read the preset from. |
| `pointer.line` | integer | The line of the preset name. |
| `poll_ms` | integer | Polled kinds: 0 is every tick, else the rate. |
| `port` | integer | udp.listen: the port. |
| `port_setting` | string | udp.listen: a setting that replaces the port. |
| `redact` | object of list of string by key | Fields replaced in recordings, by event name ("\*" for every event). |
| `replace_rumble` | boolean | game.output: while the mod plays, the game's rumble does not reach the controller, and the mod's feels play in its place. |
| `replace_rumble_setting` | string | game.output: a true or false setting of the mod that replaces replace_rumble, so the player can choose. |
| `rescan_ms` | integer | file.jsonl, file.text: re-glob at most this often. |
| `settings` | object of string by key | screen: reader settings -> the mod's setting names. |
| `sounds` | object of object by key | sound: the game's sounds it listens for, by name; each is the input &lt;id>:&lt;name>. |
| `sounds.*.cooldown_ms` | integer | No second input before this; default: the sound's length. |
| `sounds.*.min_db` | number | Quieter plays are left out, in dBFS; default -50. |
| `sounds.*.min_match` | number | How much of the clip must be found, 0.2 to 1; default 0.5. |
| `sounds.*.pitch` | integer | Semitones the game may play it higher or lower, 0 to 3; default 0. |
| `sounds.*.print` | string | Its print file in the mod: sounds/&lt;name>.print.json, made with brr sound print. |
| `sounds.*.prints` | list of string | 1 to 8 print files: variants of one sound. |
| `stale_ms` | integer | Values older than this read as absent. |
| `tick_ms` | integer | screen: the capture rate. |
| `type` | string | The kind: file.jsonl, file.json, file.text, bindings, screen, udp.listen, memory, sound, game.output. |
| `values` | object of object by key | memory: the values to read, by name. |
| `values.*.address` | string | A module and a hex offset: Game.exe+1A2B3C0. |
| `values.*.instruction_end` | integer | With relative: where the instruction ends, from the pattern's start; default: scan_offset + 4. |
| `values.*.length` | integer | A string's length in bytes, 1 to 256. |
| `values.*.module` | string | The module the pattern is in; default: the game's exe. |
| `values.*.offsets` | list of string or integer | The pointer chain, in the order it is followed. |
| `values.*.poll_ms` | integer | How often it is read; default: the sensor's poll_ms, else 100; at least 25. |
| `values.*.relative` | boolean | The 4 bytes at scan_offset are an instruction's displacement: the place is where the instruction ends plus it. |
| `values.*.scan` | string | A byte pattern with wildcards: 48 8B 05 ?? ?? ?? ?? 48 85 C0. |
| `values.*.scan_offset` | integer | Added to where the pattern starts. |
| `values.*.type` | one of u8, i32, u32, i64, f32, f64, bool, string | What it is. |
| `when` | one of game.running, window.in_front | A condition that gates the sensor. |
| `window` | string | screen: the game key whose window it captures. |

## The game

A game mod gets these without declaring them, from its manifest's `games` entry:

- inputs `game:check` (every 3 s, with `running`), `game:started` and `game:closed`;
- the snapshot `sensors.game`: `running`, and `in_front` (the game's window is the one in front).

## file.json

Reads a JSON file a game rewrites, such as a status file:

```jsonc
"status": { "type": "file.json",
            "path": "{path:saves}/status.json", "poll_ms": 0, "first_change": "silent",
            "fields": {
              "Flags":  { "type": "bits", "bits": { "Docked": 0, "Landed": 1 } },
              "Speed":  { "type": "number" },
              "Health": { "type": "number?" } } }
```

| Field | What it does |
|---|---|
| `path` | the file |
| `fields` | the fields to read, each with a `type`: `int`, `number`, `string`, `bool`, the same with `?` (missing reads as nil) or `[]` (a list), `object`, `any`, or `bits` with named bits |
| `first_change` | `silent` (default): the first read sets the values without an input; `change`: the first read is a change |
| `edges` | inputs for a change of one value, by name: `"docked": { "on": "Flags.Docked", "kind": "rise" }`, with `change`, `rise`, `fall` and `threshold`; a threshold fires as the value passes `above`, and can fire again once the value has fallen under `below` |

A read that is empty, unchanged or broken is skipped, since games rewrite such files in place. Inputs: `<id>:change` with `old` and `new`, and one per edge. Named bits are read as `st.flags.Docked`, the number as `st.Flags`.

## file.jsonl

Follows the newest file of JSON lines in a folder, such as a game's event log, and delivers each line as an input named by one of its fields:

```jsonc
"events": { "type": "file.jsonl",
            "dir": "{path:saves}", "glob": "Events.*.log", "pick": "newest_mtime",
            "name_field": "event", "catch_up": "state_only", "poll_ms": 0 }
```

| Field | What it does |
|---|---|
| `dir`, `glob` | the folder and the files to follow |
| `pick` | `newest_mtime`: the file written last |
| `rescan_ms` | how often to look for a newer file (default 2000) |
| `name_field` | the field that names the input (default `event`) |
| `catch_up` | the lines already in the file at start: `state_only` (default) delivers them with `live` false, so the mod rebuilds its state and plays nothing; `skip` leaves them out; `play` delivers them as live |

Inputs: `<id>:<name>` with the line's fields. Rules fire only on live lines unless they say `"replay": true`. Lua can ask for the newest line of a kind: `sensor.call("events", "last", "CarSetup")`.

## file.text

Follows a text log like file.jsonl does, and turns lines into inputs with Go regular expressions:

```jsonc
"gamelog": { "type": "file.text",
             "dir": "{SteamApp:12345}/logs", "glob": "Game.log", "pick": "newest_mtime", "catch_up": "skip",
             "lines": { "arrived": { "re": "<Arrived> .* at (?P<to>\\S+)" } } }
```

Inputs: `<id>:<rule>`, such as `gamelog:arrived`, with each named capture as a field (`ev.to`).

## udp.listen

Listens for a game's telemetry on a UDP port and decodes it with a [binary layout](layouts.md):

```jsonc
"telemetry": { "type": "udp.listen",
               "bind": "127.0.0.1", "port": 20777, "port_setting": "udp_port",
               "layout": "layouts/racer.layout.json",
               "forward": ["127.0.0.1:20778"], "stale_ms": 1000, "when": "game.running" }
```

| Field | What it does |
|---|---|
| `bind` | `127.0.0.1`: this PC only |
| `port` | the port; the manifest's `permissions.udp` must list it |
| `port_setting` | a setting that replaces the port, so the player can change it |
| `layout` | the layout file |
| `forward` | send each packet on, unchanged, to these addresses first, so another tool keeps getting it |

Inputs: the layout's edges and event packets, in the order the packets came, then one `<id>:change` a tick with the newest values. Forwarding off this PC needs the address in `permissions.udp_forward` and the player's consent. When another program has the port, the sensor says so and tries again.

## bindings

Reads a game's control preset, an XML file, and turns the controller's buttons and the keyboard's keys into the game's actions, with the game's modifier keys and combos. A bindings layout in the mod (`bindings/<name>.json`) says how the game writes its preset: the elements of an action, how it names the controller's buttons and sticks and the keys, and which actions the mod reacts to, under the mod's own names:

```jsonc
{
  "bindings": 1,
  "preset": "PresetName",
  "keys": ["Primary", "Secondary"],
  "modifier": "Modifier",
  "device": "Device", "key": "Key", "value": "Value",
  "actions": { "horn": "Horn", "boost": "UseBoost" },
  "keyboard": { "devices": ["Keyboard"], "prefix": "Key_" },
  "pads": [ { "devices": ["DualSense"], "buttons": { "Joy_1": "square", "Joy_2": "cross" } } ]
}
```

Every field of a bindings layout is in the [file reference](files.md#bindings). The sensor's `dir`, `file` and `pointer` say where the preset is and which file names the one in use. Inputs: `binds:change` when the preset is read again, and `binds:action` for each action that fires, with `action` (the mod's name). Rules name one as `binds:action=horn`. The sensor reads keys only while the game's window is in front, and needs the `keyboard` permission.

## screen

Reads numbers, bars, icons and lists from the game's window, where a game shows something it writes nowhere else. A screen layout in the mod (`screen/<name>.json`) describes:

- `colours`: the colours to look for, each with its standard value and how close a pixel must be;
- `fonts`: the digit shapes of the game's numbers, one file per font, made from screenshots;
- `parts`: what to read and where: a `number` ("NN%" or plain digits) in an area of the window or next to another number, a `pair` of numbers, `bars` or a `zone` next to a number, `lists` of named entries, a `bar` without a number, or an `icon`. Each part says what its changes mean, such as a drop of a number as a hit with a strength. See [Screen layouts](#screen-layouts).

The sensor's `window` names the game, `layout` the file, `settings` the mod's settings that switch it on and set the colours, and `palette` the files with the game's own colour settings and the colours it learned. The mod's Lua tells it when to read with `sensor.set`. It needs the `screen` permission and the game in a window or borderless. `.\brr -hudtest <screenshots>` reads screenshots as the sensor would. Bururu captures the window with Windows' standard screen capture; it never opens the game's process. On Linux it reads the game's X11 window, which Proton gives every game by default. A game in a Wayland window cannot be read there, and the sensor says so.

### Screen layouts

Most games show health, stamina and the like as bars without a number, some numbers without "%", and small icons. These are the parts for them. Every field of a layout is in the [file reference](files.md#screen).

**Where a part is.** `area` is `[x0, y0, x1, y1]`, each 0 to 1: parts of the window, left top first. Most games draw their HUD for a 16:9 screen, pin each element to a corner or an edge, and scale it with the window. Give a part an `anchor` (`top-left`, `top`, `top-right`, `left`, `centre`, `right`, `bottom-left`, `bottom` or `bottom-right`) and its area is in the widest 16:9 box that fits the window, with the box at that place. The same numbers then fit 1080p, 1440p and 4K, and ultrawide and 16:10 windows. Any part with an area takes an anchor.

**A bar.** How full a bar without a number is:

```jsonc
"health": { "kind": "bar", "anchor": "bottom-left", "area": [0.0313, 0.9259, 0.2396, 0.937],
            "colour": "health_red", "empty": "bar_grey",
            "events": { "drop": { "kind": "hit", "base": 0.2, "div": 50, "damage": true } } }
```

The area is the bar from end to end, cut tight. `colour` is the fill's colour, and the fill runs from `from`: `left` (the default), `right`, `top` or `bottom`. `empty` is the colour of the empty part, when it is solid and unlike the world around it: with it an empty bar reads as 0, without it an empty bar cannot be told from a hidden HUD and reads as not in view. Gaps in the fill up to `segment_gap` of the length (0.03) are bridged, for bars made of segments. Many games leave a trail behind the fill for a moment after a hit. A trail in a colour of its own never counts as fill. A darker trail in the fill's colour does not count either: fill dimmer than `trail` (0.7) of the fill's start is that trail; 0 turns this off. The value is the middle one of the last `smooth` reads (3). In Lua and rules the part is `sensors.<id>.health` with `value` (0 to 1), `ok` and `at`. A `drop` has the drop in percent of the bar as x, so `damage` is in percent too; a `rise` comes when it grows.

**A plain number.** A number without "%", such as ammo or gold:

```jsonc
"ammo": { "kind": "number", "anchor": "bottom-right", "area": [0.865, 0.917, 0.99, 0.981],
          "colour": "text", "font": "digits", "percent": false, "filter": { "max": 250 } }
```

`"percent": false` reads digits only. `filter.max` is optional here: the highest value, which also sets how many digits it may have (up to 6). Its font is made from the game's plain numbers (below). The rest works as for a "NN%".

**An icon.** Which of a few small pictures shows, such as a mount or a status effect:

```jsonc
"mount": { "kind": "icon", "anchor": "top-right", "area": [0.922, 0.037, 0.984, 0.13], "size": 0.0556,
           "images": { "horse": "screen/horse.png", "boat": "screen/boat.png" },
           "events": { "show": { "kind": "mounted", "base": 1 }, "hide": { "kind": "dismounted", "base": 1 } } }
```

Each picture is a PNG in the mod, cut from a screenshot, with the game world around the icon made transparent. The icon's own slot or frame can stay: the empty slot does not match it. A symbol of one colour must stand out from what is around it: white snow is no white symbol. `size` is the icon's height as a part of the window height. The area may be bigger than the icon, and the icon is found anywhere in it: up to 8 times its size across or down, with room for 16 icons in all (such as 4 by 4, or 8 by 2). `tolerance` (0.12) is how much the colours may differ on average, and `smooth` (2) is how many reads in a row change what shows. The part has `value` (the picture's name, `""` for none), `shown`, `ok` and `at`. `show` comes when a picture shows or another takes its place, `hide` when it goes; both have `ev.picture`.

**Measuring from a screenshot.** Take a screenshot of the game, find a region's pixels in an image editor, and run:

```powershell
.\brr mod measure shot.png 60,1000,460,1012
```

It prints the region as an `area`, the nearest `anchor` with the area for it, the region's height as an icon's `size`, its main colours, and a bar it sees: from which end, how full, and its fill and empty colours. `--icon screen\horse.png` also saves the region as a picture. For a font, measure a number and say what it shows: `--text 250 --font screen\digits.json --colour text --layout screen\hud.json` adds the digits 2, 5 and 0 to the font, makes it when it is not there, and says which digits it still needs. The reader cuts plain numbers and "NN%" apart differently, so make a plain number's font from plain numbers, and a "NN%" font from "NN%".

## memory

Reads values from the game's own memory, on Windows. Bururu only reads: it never writes to the game and never puts code into it. It does this only for a game the mod list marks as offline with no anti-cheat, see [The anti-cheat rule](#the-anti-cheat-rule).

```jsonc
"ship": { "type": "memory", "poll_ms": 100,
          "values": {
            "hull":  { "type": "f32", "scan": "48 8B 05 ?? ?? ?? ?? 48 85 C0", "scan_offset": 3, "relative": true,
                       "offsets": ["1C8", "10"] },
            "gear":  { "type": "i32", "address": "Game.exe+1A2B3C0", "poll_ms": 50 },
            "pilot": { "type": "string", "length": 32, "address": "Game.exe+1A2B400", "offsets": ["0"] } } }
```

Each value has a type and a place:

| Field | What it does |
|---|---|
| `type` | `u8`, `i32`, `u32`, `i64`, `f32`, `f64`, `bool` or `string` |
| `length` | a string's length in bytes, 1 to 256; the text ends at the first zero byte |
| `address` | a module and a hex offset, as Cheat Engine writes it: `Game.exe+1A2B3C0`; a hex offset alone counts from the game's exe |
| `scan` | a byte pattern instead of an address: hex bytes, `??` for any byte, `4?` for half a byte; at least 4 bytes without a wildcard; without spaces each byte is two characters |
| `module` | the module the pattern is in, such as `engine.dll`; the game's exe when left out |
| `scan_offset` | added to where the pattern starts |
| `relative` | the 4 bytes at `scan_offset` are the displacement of an instruction such as `mov rax,[rip+...]`: the place is where the instruction ends plus that displacement |
| `instruction_end` | with `relative`: where that instruction ends, counted from the pattern's start; `scan_offset` + 4 when left out |
| `offsets` | the pointer chain: at each offset Bururu reads the pointer where it is and adds the offset, as Cheat Engine does; hex text (`"1C8"`) or whole numbers |
| `poll_ms` | how often the value is read: the sensor's `poll_ms` when left out, else 100; at least 25 |

A pattern survives most game updates; an address usually does not. Bururu looks for the patterns when the game starts, and again for a minute while the game loads its modules. A value it cannot find, or a pointer that leads nowhere, reads as nil, and the log says once which patterns were not found. A sensor reads at most 64 values.

Inputs: `<id>:change` with `old` and `new` when a value changes, from the second read on. The snapshot `sensors.<id>` holds the latest values, such as `sensors.ship.hull`. When the game closes, the values go and Bururu lets go of the game's process.

The sensor needs the `memory` permission. The player sees "Reads the game's memory. Allowed because the game is offline with no anti-cheat." on the mod's page and in the mod list. On Linux and macOS the sensor is off: it says once that this system cannot read a game's memory, and the mod runs on its other sensors. Protected processes, a refused access and patterns that are gone after an update each give one line in the log, and the mod keeps running the same way.

### From Cheat Engine

- A pointer: the base address `"Game.exe"+01A2B3C0` becomes `"address": "Game.exe+1A2B3C0"`. Cheat Engine's pointer window lists the offsets from the value at the top down to the base address; the offset at the bottom, next to the base address, comes first. Write them from the bottom up. In the pointer scanner's results, `Offset 0`, `Offset 1` and on are already in that order. A saved table (`.CT`) lists them as the window does, so reverse its `<Offsets>`.
- An AOB scan: `aobscanmodule(hull, Game.exe, 48 8B 05 ?? ?? ?? ?? 48 85 C0)` becomes `"scan": "48 8B 05 ?? ?? ?? ?? 48 85 C0"`, with `"module"` when the module is not the game's exe. When the script then takes the address out of the instruction (it reads 4 bytes at `hull+3` and adds them to `hull+7`), that is `"scan_offset": 3, "relative": true`.
- Types: Byte is `u8`, 4 Bytes is `i32` (or `u32`), 8 Bytes is `i64`, Float is `f32`, Double is `f64`, and String is `string` with its length. Numbers reach Lua as Lua numbers, so an `i64` above 2^53 loses its last digits.

### Finding a value with brr memory

`brr memory` finds where a game keeps a value, the way Cheat Engine's scans do, and gives you the fields of a memory value for it. It only reads the game's memory. Use it on a game that plays offline with no anti-cheat, while the game runs. `scan`, `pointers`, `pattern` and `check` name the game by its exe name or its process ID (PID), and `next` and `watch` work on the game of the last scan. It refuses Windows' own processes, services and programs of another user. It runs on Windows only, and `.\brr memory help` lists every option.

An example: the hull of your ship shows 100, and you want it as a value.

1. Scan for it: `.\brr memory scan MyGame.exe f32 100`. A float is found as rounded to the digits you give, so `100` finds 99.5 to 100.5. When you do not know the number, give a range (`1..1000`) or `unknown`. An `unknown` scan keeps a packed copy of all the memory the game writes, up to 8 GB of it, and the copy can take that much room on disk; for a bigger game, give a range.
2. Let the hull take a hit, and keep the places that went down: `.\brr memory next decreased`. Then give the new number: `.\brr memory next 87.5`. `next` also takes `increased`, `changed`, `unchanged` and a range. A `next` that keeps no place leaves the scan as it was, so you can give another word or number. Go on until a few places are left. `.\brr memory watch` reads the values of the first 8 places a few times a second and prints a line each time one changes, so you see which one follows the game.
3. Find the pointer chains to the place that is left: `.\brr memory pointers MyGame.exe 1F2A3B4C5D0`. It prints them shortest first, as `"address"` and `"offsets"` in the order this sensor takes them, and keeps them. Most chains break when the game starts again: restart it, find the value again with steps 1 and 2, and run `.\brr memory pointers MyGame.exe <its new address> --keep`. Only the chains that still lead to the value stay; when none does, the file stays as it was, so a wrong address loses nothing. `--depth` (4 pointers) and `--offset` (1000, in hex) set how long a chain may be and how far past a pointer the next place may be.
4. For a value that should last through game updates, make a pattern for the chain's address: `.\brr memory pattern MyGame.exe MyGame.exe+1A2B3C0` finds the code that reads that address and prints a `"scan"` with its `"scan_offset"`, `"relative"`, `"instruction_end"` and `"module"`. Each pattern is found once in the module, and checked.
5. Put the value in your mod's `sensors.json` and check it: `.\brr memory check MyGame.exe mods\my-mod` reads each value of the mod's memory sensors as the sensor does, and prints its place, each pointer of its chain and what it reads now.

The value of this example, with its address:

```jsonc
"hull": { "type": "f32", "address": "MyGame.exe+1A2B3C0", "offsets": ["18", "48"] }
```

or with the pattern from step 4 in place of the address:

```jsonc
"hull": { "type": "f32", "scan": "48 8B 05 ?? ?? ?? ?? 48 85 C0 74 0A", "scan_offset": 3, "relative": true,
          "instruction_end": 7, "module": "MyGame.exe", "offsets": ["18", "48"] }
```

`memory check` also takes a `sensors.json` file, or a file with one value in it. The scan and the chains are kept in `%LOCALAPPDATA%\Bururu\memory`; `--file` keeps them elsewhere, such as one pointer file for each value.

### The anti-cheat rule

Bururu reads a game's memory only when the game's record in the mod list says `"anti_cheat": "none"`: the game has no anti-cheat and plays offline. Any other value (`server` for a game that plays online only, `user` or `kernel` for an anti-cheat in the game or in the system), or no value, keeps the sensor off with one line in the log. It opens only the game's own process: an exe that the mod's manifest gives for the game and the game's record in the mod list names too. A game that the mod's own manifest gives an anti-cheat (`server`, `user` or `kernel`) keeps it off too. The list's maintainers set the value after checking the game; ask for it when you list your mod. `bururu mod check` reminds you of the rule for each memory sensor.

## sound

Listens to the game's own sound while it runs, and gives an input when one of the sounds the mod knows plays. The mod knows a sound by its print: a fingerprint of a short clip of the game, made with `brr sound print`. A print holds where the clip's spectrogram has its peaks and how strong each is. It cannot be played back, so a mod holds none of the game's sounds. How to make a sound mod: [Sounds of a game](quickstart-sound.md).

```jsonc
"sounds": {
  "type": "sound",
  "sounds": {
    "pickup": { "print": "sounds/pickup.print.json" },
    "reload": { "prints": ["sounds/reload-a.print.json", "sounds/reload-b.print.json"] },
    "alarm":  { "print": "sounds/alarm.print.json", "min_match": 0.7, "cooldown_ms": 1000 }
  }
}
```

| Field | Default | What it does |
|---|---|---|
| `sounds` | | the sounds, up to 32, by name; a name is a word of letters, digits and `_` |
| `print`, `prints` | | the print file, or up to 8 (variants of one sound); 64 print files in a mod at most |
| `min_match` | 0.5 | how much of the clip must be found, 0.2 to 1; higher gives fewer wrong finds and misses more |
| `pitch` | 0 | semitones the game may play it higher or lower, 0 to 3, as games do to vary a sound |
| `cooldown_ms` | the sound's length | no second input before this |
| `min_db` | -50 | plays quieter than this (dBFS) are left out |

Inputs: `<id>:<name>`, such as `sounds:pickup`, with `match` (0 to 1), `strength` (0 to 1: how loud it played against the clip, for a rule's `scale`), `level_db`, `pitch` (the semitones it played at) and `late_ms` (from the sound's start to the input). The snapshot `sensors.<id>` has `listening` (the game's sound comes) and `last` (the name of the sound heard last). A rule:

```jsonc
"pickup": { "on": "sounds:pickup", "play": { "feel": "thump", "scale": "ev.strength" } }
```

What to know:

- **Which sounds work.** A sound with a clear pitch or several parts works best: a pickup chime, a reload, an alarm, a menu sound, a low-health warning. A noise burst, a very short tick and a steady hum have few peaks that stay the same, and are hard to find; `brr sound print` says so. A sound made of notes can be taken for the game's music when the music plays the same notes: try every print on a long recording of the game's music with `brr sound find`, and raise `min_match` when it finds the sound where it is not.
- **The delay.** The input comes about 100 ms after the sound starts, sometimes more, as the sensor must hear the start of it, then up to a tick more. A feel then still has the way through the haptics stream to go. On the cable the haptics play through a shorter buffer while the sensor listens, and the log says once how long a push takes there (`Native haptics: through WASAPI a push takes about ...`); on Bluetooth it is 10 to 40 ms. That suits pickups, reloads, alarms and menus. For the first moment of a shot or an explosion, it is late.
- **What it listens to.** Only the game's own programs (the game's exe and the programs it started), through Windows' process loopback, and only while the game runs and the effects are not paused. Never the whole PC and never the microphone. It also refuses the programs on Bururu's list of those it never listens to, whatever a mod says: browsers, voice and chat apps, music and video players, streaming, recording and remote-play tools, launchers and Windows' own. Bururu looks at the sound as it comes and keeps nothing; nothing is sent.
- **Which game.** An exe that the mod's manifest gives for the game and the game's record in the mod list names too. When the manifest's `markers` name files of the game's folder, only a program that runs from such a folder. For a game the list names no program of yet, with **Developer tools** on, the manifest's exe is enough where one of its `markers` is in the exe's folder or above it, and the log says so. A marker that is the program itself does not count there, as every program of that name has it: name a file or folder of the game, such as its data folder.
- **Windows 11.** Process loopback needs Windows 11. On Windows 10 the sensor is off with one line in the log, and the mod runs on its other sensors. On Linux and macOS it is off too.

The sensor needs the `sound` permission. The player sees "Listens to the game's own sound while it runs. Nothing is kept or sent." in the install dialog and on the mod's page.

## game.output

Reads what the game sends to its controller: the rumble, the trigger effects and the lightbar. It works on Windows, for a game that plays on Bururu's virtual controller. A game that reads the DualSense itself sends Bururu nothing, and the sensor waits. It needs no permission: it reads only what the game sends to Bururu's own controller.

```jsonc
"out": { "type": "game.output", "replace_rumble": true }
```

| Field | What it does |
|---|---|
| `replace_rumble` | `true`: while the mod plays, the game's rumble does not reach the controller, and the mod's own feels play in its place. The game's trigger effects and lightbar still pass. Default `false` |
| `replace_rumble_setting` | a setting of the mod with `"type": "bool"` that replaces `replace_rumble`, so the player can choose between the game's rumble and the mod's feels |

The snapshot `sensors.<id>`:

| Field | What it is |
|---|---|
| `rumble` | `true` while one of the game's motors runs |
| `left`, `right` | the left (heavy) and the right (light) motor, 0 to 1, less the motor power the game turned down |
| `kind` | how the game rumbles while `rumble` is true, `older` or `newer`, else `""`. The older kind plays about twice as strong as the newer one at the same level |
| `lightbar` | a list of three numbers, red, green and blue, 0 to 255 each, once the game set a colour |
| `l2_effect`, `r2_effect` | the trigger effect the game set, once it set one: 11 numbers as the controller reads them, the first is the effect's mode |

Inputs: `<id>:change` with `old` and `new`, each time the game sends something that changes the snapshot. A game sends its rumble about every 10 to 20 ms while it rumbles, so a tick (25 ms) can bring two or three changes, in order.

The mod takes the game's rumble over only while it plays: the game runs, the mod says the player is in control (`active`), the effects are not paused and the haptics are on, the sensor's `when` holds, and none of the mod's Lua handlers was switched off for failing. Otherwise, such as in a menu the mod does not mark active, the game's rumble reaches the controller as before.

A recording keeps each change as an event with the new snapshot, and a replay gives it again as a change input. Lines written by hand may leave out what is at rest; `rumble` follows the motors, and `kind` is `older` unless the line says `newer`:

```
{"t_ms": 1000, "sensor": "out", "events": [{"left": 0.6, "right": 0.2}]}
{"t_ms": 1800, "sensor": "out", "events": [{}]}
```

The game sends this only to the running Bururu, so the **Developer** tab of the **Mods** page records it, and `mod record` leaves it out.

An example: the game's rumble as a smooth swell on the actuators. The feel, `feels/rumble_swell.feel.json`:

```jsonc
{ "feel": 1, "name": "rumble_swell", "kind": "layer",
  "voices": [ { "id": "body", "type": "synth", "wave": "Sine", "f0_hz": 170, "level": 0.35 } ] }
```

and in the mod's Lua:

```lua
bururu.on_tick(function(t)
  local out = sensors.out
  if out and out.rumble then
    feel.layer("rumble", "rumble_swell", gomath.max(out.left, out.right))
  end
end)
```

A layer glides to each new level and fades out once the rumble stops. Some games rumble for seconds at a time; for those, short swells or pushes (`feel.play`) feel better than a held tone. Below about 100 Hz the actuators rattle.

## The controller and the keyboard

The controller is no sensor a mod declares. Bururu reads it once per active tick and hands it to Lua as `t.pad`: the triggers, the buttons held and pressed (`south east west north l1 r1 ...`, also by their Sony and Xbox names), the sticks and the gyro. The fields: [Pad](lua-api.md#pad). A mod never reads the keyboard; the bindings sensor delivers actions.

## Status

Each sensor shows its state on the **Mods** page:

| State | Meaning |
|---|---|
| `ok` | data arrives |
| `waiting` | no data yet, or its `when` does not hold |
| `error` | a missing file, a busy port, or Windows did not share the game's sound; the text says which |
| `unavailable` | its path does not exist on this PC, or it does not run here (the memory sensor off Windows, or for a game it may not read; the sound sensor before Windows 11, or for an exe the mod list does not name; the game output sensor off Windows) |

The input `sensor:status` comes when a sensor's state changes, so a mod can say on its page that telemetry is off. Only the memory sensor reads a game's memory, and only as [memory](#memory) says.
