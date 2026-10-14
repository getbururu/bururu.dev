# Mod file reference

Every field of every mod file, from the same schemas `bururu mod schema` writes into a mod's `.bururu/schema/` folder for an editor. Every file may hold `//` and `/* */` comments, trailing commas, and keys that start with `$` or `_comment`, which are left out. An unknown field is an error with a "did you mean" hint.

How to read a path: `a.b` is the field `b` inside `a`; `a.*.b` is the field `b` of each entry of `a`, which is keyed by id or by name; `a[].b` is the field `b` of each item of the list `a`. "A number or a Lua expression" takes either; a text field marked "a literal, or {"expr": ...}" takes a computed value as `{"expr": "..."}`.

| File kind | Files |
|---|---|
| [manifest](#manifest) | `manifest.json` |
| [sensors](#sensors) | `sensors.json` |
| [rules](#rules) | `rules/*.rules.json`, `overrides/*/rules/*.rules.json`, `final/*/rules/*.rules.json` |
| [lights](#lights) | `lights.json`, `overrides/*/lights.json`, `final/*/lights.json` |
| [feel](#feel) | `feels/*.feel.json`, `overrides/*/feels/*.feel.json`, `final/*/feels/*.feel.json` |
| [settings](#settings) | `settings.schema.json` |
| [layout](#layout) | `layouts/*.layout.json`, `overrides/*/layouts/*.layout.json` |
| [screen](#screen) | `screen/*.json` |
| [bindings](#bindings) | `bindings/*.json` |
| [migration](#migration) | `migrations/*.json` |
| [vocab](#vocab) | `vocab/*.vocab.json` |
| [soundprint](#soundprint) | `sounds/*.print.json` |

## manifest

A mod's manifest.json. What the mod is: its id, version, kind, games, paths, dependencies, outputs and permissions. See [the mod package](index.md#the-mod-package).

| Field | Type | What it is |
|---|---|---|
| `api` | string | The mod API it is made for: "0.12". Mods made for "0.10" and "0.11" still load; macos entries and the game.output sensor need "0.12". |
| `authors` | list of string | Shown on the Mods page. |
| `demo` | string | The replay "Play demo" plays: demo/demo.replay.jsonl. |
| `dependencies` | object of object by key | Dependencies by mod id. |
| `dependencies.*.kind` | one of required, optional, incompatible | required (default), optional or incompatible. |
| `dependencies.*.order` | one of after, before, none | Load after (default), before, or no order. |
| `dependencies.*.version` | string | A range: ">=1.0.0 &lt;2.0.0". |
| `dependencies.*.why` | string | Incompatible: the reason the Mods page shows. |
| `entry` | string | The Lua entry, lua/main.lua; none: no Lua of its own. |
| `feels` | object | Feel options. |
| `feels.rumble_default` | one of none, derived | What feels without a rumble entry do on the motors. |
| `for` | list of string | Add-ons: the mod ids and vocabularies it changes. |
| `games` | object of object by key | Game mods: the games, by key. |
| `games.*.anti_cheat` | object | The game's anti-cheat. |
| `games.*.anti_cheat.kind` | one of none, server, user, kernel | none, server, user or kernel. |
| `games.*.anti_cheat.notes` | string | Notes for people. |
| `games.*.anti_cheat.verified_safe` | list of one of screen, keyboard | With kernel: sensors known to be safe. |
| `games.*.markers` | list of string | Files that prove an install folder, relative to it with forward slashes; '\*' in one part. |
| `games.*.name` | string | The game's name in words and logs. |
| `games.*.native_dualsense` | boolean | The game drives a DualSense itself. |
| `games.*.native_hint` | string | Where the game's own vibration and trigger switches are. |
| `games.*.process` | object of object by key | Its executables per OS (windows, linux, macos). On Linux and macOS the Windows ones stand in when that OS lists none. On macOS the program is the one in the app's Contents/MacOS. |
| `games.*.process.*.game` | list of string | The game's executables. |
| `games.*.process.*.launcher` | list of string | Its launcher's executables. |
| `games.*.process.*.match` | one of comm, cmdline | Not used: on Linux a name matches the process name and the exe path alike. |
| `games.*.short_name` | string | The short name in backend words; default: the name. |
| `games.*.steam_app` | integer | Its Steam app id. On Linux it gives the window's WM_CLASS and the Proton prefix of the paths. |
| `games.*.stores` | object | Its ids in each store, so the game scan finds it before the mod is listed. |
| `games.*.stores.battlenet` | list of string | Battle.net agent uids. |
| `games.*.stores.ea` | list of string | EA app softwareIds; matched through exes and markers only. |
| `games.*.stores.epic` | list of object | Epic ids: app, else namespace and item. |
| `games.*.stores.epic[].app` | string | The AppName. |
| `games.*.stores.epic[].item` | string | The CatalogItemId. |
| `games.*.stores.epic[].namespace` | string | The CatalogNamespace. |
| `games.*.stores.gog` | list of integer | GOG product ids. |
| `games.*.stores.steam` | list of integer | Steam app ids. |
| `games.*.stores.ubisoft` | list of integer | Ubisoft Connect product ids. |
| `games.*.stores.xbox` | list of object | Xbox app ids. |
| `games.*.stores.xbox[].family` | string | The package family name. |
| `games.*.stores.xbox[].store_id` | string | The Microsoft Store id. |
| `games.*.virtual_pad` | one of never, shared, game, launcher_or_game | The controller the game sees unless the mod list has a profile for it or the player picks one in Bururu. game: Bururu's virtual DualShock 4; launcher_or_game: the same, plugged in when the launcher starts; never: the DualSense itself; shared: for a program that reads no controller (Bururu changes nothing about the controller for it, and the player gets no choice). |
| `games.*.window` | object of object by key | Its window per OS (windows, linux, macos). |
| `games.*.window.*.bundle_id` | string | The macOS app's bundle id: com.studio.game. |
| `games.*.window.*.class` | string | The Windows window class. |
| `games.*.window.*.wm_class` | string | The Linux WM_CLASS; default: steam_app_&lt;steam_app>, as Proton names it. |
| `homepage` | string | An https address, opened from the Mods page. |
| `hooks` | object of object by key | Extension points it runs with hook.run, name@version. |
| `hooks.*.doc` | string | What a handler may change. |
| `icon` | string | A name from the window's icon set. |
| `id` | string | The id: a lower-case letter, then 1-39 lower-case letters, digits or '-'. Never changes. |
| `kind` | one of game, addon, library | What the mod is. |
| `license` | string | Shown on the Mods page. |
| `load_after` | list of string | Mods it loads after, without depending on them. |
| `load_before` | list of string | Mods it loads before. |
| `manifest` | integer | The manifest format: 1. |
| `name` | string | The name the Mods page shows. |
| `os` | list of one of windows, linux, macos | Where it works; none: everywhere. |
| `outputs` | list of one of haptics, rumble, triggers, lightbar, player_leds, mic, gyro_hold | Game mods: the output groups it drives; anything else is dropped. |
| `paths` | object of object by key | Declared folders and files, by name. |
| `paths.*.label` | string | Its name in messages. |
| `paths.*.linux` | string | Replaces path on Linux. Its Windows folders, such as {LocalAppData}, are the Linux ones unless proton is set. |
| `paths.*.macos` | string | Replaces path on macOS. There {LocalAppData} and {RoamingAppData} are ~/Library/Application Support. |
| `paths.*.path` | string | The path with tokens: {SavedGames}/... |
| `paths.*.proton` | integer | A Steam app id: on Linux the path resolves inside its Proton prefix. Default: the steam_app of the mod's game, for a path without a linux form. |
| `paths.*.setting` | string | A path setting that replaces it when not empty. |
| `paths.*.windows` | string | Replaces path on Windows. |
| `permissions` | object | What it may reach. |
| `permissions.files` | list of string | Declared path names it reads. |
| `permissions.keyboard` | boolean | It reads the keyboard (through bindings). |
| `permissions.memory` | boolean | It reads the game's memory (the memory sensor). |
| `permissions.screen` | boolean | It reads the screen. |
| `permissions.sound` | boolean | It listens to the game's own sound while it runs (the sound sensor). |
| `permissions.udp` | list of integer | UDP ports it listens on. |
| `permissions.udp_forward` | list of string | Forward targets off the loopback. |
| `provides` | list of string | Vocabularies it provides or publishes channels for: racing@1. |
| `rule_env` | object of string by key | Short names for rule expressions: sensors.&lt;id> or a Lua module (lib). |
| `summary` | string | One line for the Mods page. |
| `version` | string | A semantic version: 1.0.0. |

## sensors

A mod's sensors.json. What the mod reads. Each kind adds its own fields; see [Sensors](sensors.md).

| Field | Type | What it is |
|---|---|---|
| `items` | object of object by key | Sensor declarations by id. |
| `items.*.bind` | string | udp.listen: the address to bind. |
| `items.*.catch_up` | one of state_only, skip, play | file.jsonl, file.text: what the first file's lines do. |
| `items.*.debug_dir` | string | screen: where debug captures go. |
| `items.*.dir` | string | file.jsonl, file.text, bindings: the folder, with tokens. |
| `items.*.edges` | object of object by key | file.json: inputs when a value moves, by name. |
| `items.*.edges.*.above` | number | threshold: fires as the value passes above this. |
| `items.*.edges.*.below` | number | threshold: armed again below this; default: above. |
| `items.*.edges.*.kind` | one of change, rise, fall, threshold | change: any change; rise: from 0 or false to anything else; fall: back to 0 or false; threshold: it passes above. |
| `items.*.edges.*.on` | string | A field, or one of its named bits: Flags.Docked. |
| `items.*.fallback_file` | string | bindings: the file to try next, with {preset}. |
| `items.*.fields` | object of object by key | file.json: typed fields and named bits. |
| `items.*.fields.*.bits` | object of integer by key | bits: named bits, name -> bit index. |
| `items.*.fields.*.type` | one of bits, int, int?, int[], number, number?, number[], string, string?, string[], bool, bool?, object, any | Its type; with ? a missing field reads as nil, with [] it is a list. |
| `items.*.file` | string | bindings: the preset file, with {preset}, {major} and {minor}. |
| `items.*.first_change` | one of silent, change | file.json: the first good read sets the snapshot without a change. |
| `items.*.forward` | list of string | udp.listen: forward each datagram to these targets. |
| `items.*.glob` | string | file.jsonl, file.text: the files to follow. |
| `items.*.layout` | string | udp.listen, bindings, screen: the layout file in the mod. |
| `items.*.lines` | object of object by key | file.text: line rules with Go regular expressions. |
| `items.*.lines.*.re` | string | A Go regular expression with named captures. |
| `items.*.name_field` | string | file.jsonl: the field that names the event. |
| `items.*.order` | integer | Its place among the sensors that read at the same time. |
| `items.*.palette` | object | screen: the palette files. |
| `items.*.palette.learned_file` | string | The learned palette file. |
| `items.*.palette.matrix_file` | string | The colour matrix file. |
| `items.*.patchable` | list of string | Parameters other mods may patch. |
| `items.*.path` | string | file.json: the file, with tokens. |
| `items.*.pick` | one of newest_mtime | file.jsonl, file.text: which file. |
| `items.*.pointer` | object | bindings: where the preset name is. |
| `items.*.pointer.fallback_line` | integer | The line to try next. |
| `items.*.pointer.files` | list of string | Files to read the preset from. |
| `items.*.pointer.line` | integer | The line of the preset name. |
| `items.*.poll_ms` | integer | Polled kinds: 0 is every tick, else the rate. |
| `items.*.port` | integer | udp.listen: the port. |
| `items.*.port_setting` | string | udp.listen: a setting that replaces the port. |
| `items.*.redact` | object of list of string by key | Fields replaced in recordings, by event name ("\*" for every event). |
| `items.*.replace_rumble` | boolean | game.output: while the mod plays, the game's rumble does not reach the controller, and the mod's feels play in its place. |
| `items.*.replace_rumble_setting` | string | game.output: a true or false setting of the mod that replaces replace_rumble, so the player can choose. |
| `items.*.rescan_ms` | integer | file.jsonl, file.text: re-glob at most this often. |
| `items.*.settings` | object of string by key | screen: reader settings -> the mod's setting names. |
| `items.*.sounds` | object of object by key | sound: the game's sounds it listens for, by name; each is the input &lt;id>:&lt;name>. |
| `items.*.sounds.*.cooldown_ms` | integer | No second input before this; default: the sound's length. |
| `items.*.sounds.*.min_db` | number | Quieter plays are left out, in dBFS; default -50. |
| `items.*.sounds.*.min_match` | number | How much of the clip must be found, 0.2 to 1; default 0.5. |
| `items.*.sounds.*.pitch` | integer | Semitones the game may play it higher or lower, 0 to 3; default 0. |
| `items.*.sounds.*.print` | string | Its print file in the mod: sounds/&lt;name>.print.json, made with brr sound print. |
| `items.*.sounds.*.prints` | list of string | 1 to 8 print files: variants of one sound. |
| `items.*.stale_ms` | integer | Values older than this read as absent. |
| `items.*.tick_ms` | integer | screen: the capture rate. |
| `items.*.type` | string | The kind: file.jsonl, file.json, file.text, bindings, screen, udp.listen, memory, sound, game.output. |
| `items.*.values` | object of object by key | memory: the values to read, by name. |
| `items.*.values.*.address` | string | A module and a hex offset: Game.exe+1A2B3C0. |
| `items.*.values.*.instruction_end` | integer | With relative: where the instruction ends, from the pattern's start; default: scan_offset + 4. |
| `items.*.values.*.length` | integer | A string's length in bytes, 1 to 256. |
| `items.*.values.*.module` | string | The module the pattern is in; default: the game's exe. |
| `items.*.values.*.offsets` | list of string or integer | The pointer chain, in the order it is followed. |
| `items.*.values.*.poll_ms` | integer | How often it is read; default: the sensor's poll_ms, else 100; at least 25. |
| `items.*.values.*.relative` | boolean | The 4 bytes at scan_offset are an instruction's displacement: the place is where the instruction ends plus it. |
| `items.*.values.*.scan` | string | A byte pattern with wildcards: 48 8B 05 ?? ?? ?? ?? 48 85 C0. |
| `items.*.values.*.scan_offset` | integer | Added to where the pattern starts. |
| `items.*.values.*.type` | one of u8, i32, u32, i64, f32, f64, bool, string | What it is. |
| `items.*.when` | one of game.running, window.in_front | A condition that gates the sensor. |
| `items.*.window` | string | screen: the game key whose window it captures. |
| `sensors` | integer | The format: 1. |

## rules

A rules file. Event and edge rules and layer rules. See [Rules](rules.md).

| Field | Type | What it is |
|---|---|---|
| `defaults` | object | Defaults for every rule of this file. |
| `defaults.if` | string | ANDed with each rule's own if (a Lua expression) |
| `format` | integer | The format: 1. |
| `layers` | object of object by key | Layer rules, by id. |
| `layers.*.feel` | string or object | The feel (a name, or {"expr": "...", "one_of": [...]}) |
| `layers.*.feel.expr` | string | the feel's name (a Lua expression) |
| `layers.*.feel.one_of` | list of string | The feels the expression may name. |
| `layers.*.gain` | string or object | The gain table key (a literal, or {"expr": "..."}) |
| `layers.*.gain.expr` | string | the value (a Lua expression) |
| `layers.*.if` | string | The condition (a Lua expression) |
| `layers.*.key` | string | The synth layer key (default: the rule id). |
| `layers.*.level` | number or string | The level, 0..1 (a number or a Lua expression) |
| `layers.*.order` | number | Order inside the stage. |
| `layers.*.params` | object of number or string by key | The feel's params. |
| `layers.*.rumble` | string or boolean | The rumble table key, false for nothing on the motors. |
| `layers.*.side` | one of both, left, right or object | both, left or right (a literal, or {"expr": "..."}) |
| `layers.*.side.expr` | string | the value (a Lua expression) |
| `layers.*.stage` | string | The stage it runs in. |
| `rules` | object of object by key | Event and edge rules, by id. |
| `rules.*.cooldown_ms` | number | Does not fire again while now - last &lt;= cooldown_ms. |
| `rules.*.do` | list of object | Actions in order. |
| `rules.*.do[].channel` | object of string by key | Publish vocabulary channels. |
| `rules.*.do[].log` | string | A log line; {expr} parts are expressions. |
| `rules.*.do[].notice` | string | A toast. |
| `rules.*.do[].overlay` | string | Start an overlay of lights.json. |
| `rules.*.do[].overlay_stop` | string | Stop an overlay. |
| `rules.*.do[].play` | string or object | A feel name, or how it plays. |
| `rules.*.do[].play.amount` | number or string | What the feel's amount-scaled voices scale by (a number or a Lua expression) |
| `rules.*.do[].play.delay_ms` | number or string | Start this much later (a number or a Lua expression) |
| `rules.*.do[].play.feel` | string or object | The feel (a name, or {"expr": "...", "one_of": [...]}) |
| `rules.*.do[].play.feel.expr` | string | the feel's name (a Lua expression) |
| `rules.*.do[].play.feel.one_of` | list of string | The feels the expression may name. |
| `rules.*.do[].play.gain` | string or object | The gain table key (a literal, or {"expr": "..."}) |
| `rules.*.do[].play.gain.expr` | string | the value (a Lua expression) |
| `rules.*.do[].play.pan` | number or string | -1 left .. 1 right (a number or a Lua expression) |
| `rules.*.do[].play.params` | object of number or string by key | The feel's params. |
| `rules.*.do[].play.rumble` | string or object | The rumble table key (a literal, or {"expr": "..."}) |
| `rules.*.do[].play.rumble.expr` | string | the value (a Lua expression) |
| `rules.*.do[].play.scale` | number or string | Strength; 0 or less plays nothing (a number or a Lua expression) |
| `rules.*.do[].play.side` | one of both, left, right or object | both, left or right (a literal, or {"expr": "..."}) |
| `rules.*.do[].play.side.expr` | string | the value (a Lua expression) |
| `rules.*.do[].set` | object of string by key | Set state.&lt;path> to an expression's value. |
| `rules.*.group` | string | Among matching rules of a group only the highest priority fires. |
| `rules.*.if` | string | The condition; with defaults.if both must hold (a Lua expression) |
| `rules.*.on` | string | The input: &lt;sensor>:&lt;name>, &lt;sensor>:\*, &lt;sensor>:change, binds:action=&lt;name>. |
| `rules.*.order` | number | Sort key among the rules of one input (default 1000). |
| `rules.*.play` | string or object | A feel name, or how it plays. |
| `rules.*.play.amount` | number or string | What the feel's amount-scaled voices scale by (a number or a Lua expression) |
| `rules.*.play.delay_ms` | number or string | Start this much later (a number or a Lua expression) |
| `rules.*.play.feel` | string or object | The feel (a name, or {"expr": "...", "one_of": [...]}) |
| `rules.*.play.feel.expr` | string | the feel's name (a Lua expression) |
| `rules.*.play.feel.one_of` | list of string | The feels the expression may name. |
| `rules.*.play.gain` | string or object | The gain table key (a literal, or {"expr": "..."}) |
| `rules.*.play.gain.expr` | string | the value (a Lua expression) |
| `rules.*.play.pan` | number or string | -1 left .. 1 right (a number or a Lua expression) |
| `rules.*.play.params` | object of number or string by key | The feel's params. |
| `rules.*.play.rumble` | string or object | The rumble table key (a literal, or {"expr": "..."}) |
| `rules.*.play.rumble.expr` | string | the value (a Lua expression) |
| `rules.*.play.scale` | number or string | Strength; 0 or less plays nothing (a number or a Lua expression) |
| `rules.*.play.side` | one of both, left, right or object | both, left or right (a literal, or {"expr": "..."}) |
| `rules.*.play.side.expr` | string | the value (a Lua expression) |
| `rules.*.priority` | number | Priority within the group (default 100). |
| `rules.*.replay` | boolean | Also fires while a file source catches up (state-only rules). |
| `rules.*.stop` | boolean | No later rule sees this input. |
| `stages` | object of object by key | Layer stages, by name. |
| `stages.*.order` | number | Its place among the stages. |

## lights

A lights.json. The situation tables and the timed overlays. See [Situation tables](rules.md#situation-tables).

| Field | Type | What it is |
|---|---|---|
| `format` | integer | The format: 1. |
| `leds` | object of object by key | The player LEDs table, rows by id. |
| `leds.*.count` | number or string | 0 off, 1-5 lit (a number or a Lua expression) |
| `leds.*.final` | boolean | While the row holds, no overlay shows. |
| `leds.*.if` | string | The condition (a Lua expression) |
| `leds.*.priority` | number | Of the rows whose if holds and that set a field, the highest priority wins it. |
| `lightbar` | object of object by key | The lightbar table, rows by id. |
| `lightbar.*.blink_hz` | number or string | Blink at this rate (a number or a Lua expression) |
| `lightbar.*.bright` | number or string | A fraction of the player's brightness (a number or a Lua expression) |
| `lightbar.*.bright_abs` | number or string | An absolute brightness, 0-255 (a number or a Lua expression) |
| `lightbar.*.color` | string or object | A colour table key; it also sets the key the window shows (a literal, or {"expr": "..."}) |
| `lightbar.*.color.expr` | string | the value (a Lua expression) |
| `lightbar.*.final` | boolean | While the row holds, no overlay shows. |
| `lightbar.*.if` | string | The condition (a Lua expression) |
| `lightbar.*.key` | string or object | With color, rgb or off: the key the window shows (a literal, or {"expr": "..."}) |
| `lightbar.*.key.expr` | string | the value (a Lua expression) |
| `lightbar.*.off` | boolean | Dark. |
| `lightbar.*.priority` | number | Of the rows whose if holds and that set a field, the highest priority wins it. |
| `lightbar.*.rgb` | list of integer or string | A colour: [r, g, b] of 0-255, or a Lua expression giving {r, g, b}. |
| `mic` | object of object by key | The mic LED table, rows by id. |
| `mic.*.final` | boolean | While the row holds, no overlay shows. |
| `mic.*.if` | string | The condition (a Lua expression) |
| `mic.*.mic` | one of on, pulse, off or object | The mic LED: on, pulse or off (a literal, or {"expr": "..."}) |
| `mic.*.mic.expr` | string | the value (a Lua expression) |
| `mic.*.priority` | number | Of the rows whose if holds and that set a field, the highest priority wins it. |
| `overlays` | object of object by key | Timed overlays by id. |
| `overlays.*.blink_hz` | number | Blink at this rate. |
| `overlays.*.color` | string | lightbar: a colour table key. |
| `overlays.*.if` | string | It shows only while this holds (a Lua expression) |
| `overlays.*.layer` | one of lightbar, triggers | What it covers. |
| `overlays.*.ms` | number | How long it shows. |
| `overlays.*.priority` | number | Among running overlays the highest wins; ties: the latest start. |
| `overlays.*.trigger` | string | triggers: a trigger table key for both triggers. |
| `triggers` | object of object by key | The triggers table, rows by id. |
| `triggers.*.both` | string or object | Both triggers, with no l2 or r2 beside it: a trigger table key, or none (a literal, or {"expr": "..."}) |
| `triggers.*.both.expr` | string | the value (a Lua expression) |
| `triggers.*.final` | boolean | While the row holds, no overlay shows. |
| `triggers.*.if` | string | The condition (a Lua expression) |
| `triggers.*.l2` | string or object | The L2 trigger: a trigger table key, or none (a literal, or {"expr": "..."}) |
| `triggers.*.l2.expr` | string | the value (a Lua expression) |
| `triggers.*.l2_raw` | object | The L2 trigger's raw effect, in place of l2. |
| `triggers.*.l2_raw.mode` | one of OFF, FEEDBACK, WEAPON, VIBRATION, SLOPE_FEEDBACK, MULTIPLE_POSITION_FEEDBACK, MULTIPLE_POSITION_VIBRATION | The trigger mode. |
| `triggers.*.l2_raw.params` | list of integer | The mode's values, each in its range. |
| `triggers.*.priority` | number | Of the rows whose if holds and that set a field, the highest priority wins it. |
| `triggers.*.r2` | string or object | The R2 trigger: a trigger table key, or none (a literal, or {"expr": "..."}) |
| `triggers.*.r2.expr` | string | the value (a Lua expression) |
| `triggers.*.r2_raw` | object | The R2 trigger's raw effect, in place of r2. |
| `triggers.*.r2_raw.mode` | one of OFF, FEEDBACK, WEAPON, VIBRATION, SLOPE_FEEDBACK, MULTIPLE_POSITION_FEEDBACK, MULTIPLE_POSITION_VIBRATION | The trigger mode. |
| `triggers.*.r2_raw.params` | list of integer | The mode's values, each in its range. |

## feel

A feel. Its voices, params, mix and motor lane. See [Feels](feels.md).

| Field | Type | What it is |
|---|---|---|
| `feel` | integer | The format: 1. |
| `kind` | one of shot, layer | shot plays once; layer is held. |
| `level_mod` | list of a mod | Modulations of a layer's level. |
| `meta` | object | Free text. |
| `meta.description` | string | What it feels like. |
| `meta.editor` | string | The tool that wrote it. |
| `meta.source` | string | Who owns the voices: reaper:&lt;file>#&lt;track> or composer. |
| `meta.tags` | list of string | Tags. |
| `mix` | object | How it mixes. |
| `mix.duck` | object | A one-shot that ducks a bus as it starts. |
| `mix.duck.bus` | string | The bus. |
| `mix.duck.ms` | number | How long. |
| `mix.gain` | string | The gain table key; default: the feel's name. |
| `name` | string | The name rules and scripts play; default: the file name. |
| `params` | object of object by key | Named inputs. |
| `params.*.default` | number | The value without a setting. |
| `params.*.max` | number | The highest value. |
| `params.*.min` | number | The lowest value. |
| `rumble` | object | The motor lane. |
| `rumble.count` | integer | repeat: pulses in all. |
| `rumble.gap_ms` | number | repeat and sweep: from one pulse's start to the next. |
| `rumble.left` | number | Default left motor strength. |
| `rumble.ms` | number | Default pulse length. |
| `rumble.pattern` | one of once, repeat, sweep | How a one-shot pulses the motors. |
| `rumble.right` | number | Default right motor strength. |
| `rumble.sides` | list of one of left, right | sweep: each pulse's side. |
| `rumble.triggers` | one of l2, r2, both | A one-shot's pulses also play in these triggers, on a controller with trigger motors. |
| `voices` | list of object | The voices, summed in order. |
| `voices[].amount` | object | The level grows with a play's amount, such as a hit's damage. |
| `voices[].amount.level_max` | number | The highest level. |
| `voices[].amount.level_per` | number | Level added per amount. |
| `voices[].amount.ref` | number | The amount the voice is designed for. |
| `voices[].attack_curve` | number | Attack curve. |
| `voices[].attack_ms` | number | Attack time. |
| `voices[].band_hz` | number | Noise band width. |
| `voices[].decay_curve` | number | Decay curve. |
| `voices[].decay_ms` | number | Decay time. |
| `voices[].delay_ms` | number | Start delay. |
| `voices[].f0_hz` | number | Start frequency. |
| `voices[].f1_hz` | number | End frequency (0: none). |
| `voices[].gate_duty` | number | Gate duty. |
| `voices[].gate_hz` | number | Gate rate. |
| `voices[].grain_cycles` | number | Cycles per grain. |
| `voices[].grain_jitter` | number | Grain jitter. |
| `voices[].grain_rate` | number | Grains per second. |
| `voices[].grain_spread` | number | Grain spread. |
| `voices[].hold_ms` | number | Hold time. |
| `voices[].id` | string | A stable id; patches address voices by id. |
| `voices[].left` | number | Left actuator gain. |
| `voices[].level` | number | Level. |
| `voices[].mod` | list of a mod | Modulations of this voice. |
| `voices[].release_curve` | number | Release curve. |
| `voices[].release_ms` | number | Release time. |
| `voices[].right` | number | Right actuator gain. |
| `voices[].seed` | number | The voice's random seed. |
| `voices[].sustain` | number | Sustain level. |
| `voices[].sustain_ms` | number | Sustain time. |
| `voices[].sweep` | one of exp, linear | The sweep shape. |
| `voices[].sweep_ms` | number | Sweep time. |
| `voices[].trem_depth` | number | Tremolo depth. |
| `voices[].trem_hz` | number | Tremolo rate. |
| `voices[].type` | one of synth, curve, clip | The voice kind; a reader skips unknown kinds. |
| `voices[].wave` | one of Sine, Square, Triangle, Saw, Noise, NormNoise, BandNoise, Grains, Metal | The waveform; default: Sine. |

### A modulation (`mod`), used by `voices[].mod` and `level_mod`

A modulation.

| Field | Type | What it is |
|---|---|---|
| `add` | number | target = target + add \* source. |
| `clamp` | list of number | [lo, hi] for the result. |
| `curve` | one of linear, smooth | Between points. |
| `extend` | boolean | Extend past the last points. |
| `map` | list of list of number | A piecewise map of [x, y] points. |
| `mul` | number | target = target \* (mul \* source). |
| `op` | one of add, mul, set | How the map applies. |
| `source` | string or object | A param, a built-in (level, scale, amount, t) or an LFO. |
| `source.clock` | one of feel, wall | The feel's own time or the wall clock. |
| `source.hz` | number | Its rate. |
| `source.lfo` | one of sine, triangle, square, random | The wave. |
| `source.phase` | number | Its phase, 0..1. |
| `target` | one of level, f0_hz, f1_hz, trem_hz, trem_depth, gate_hz, gate_duty, band_hz, grain_rate, grain_spread, left, right, time | The voice field it changes. |

## settings

A settings.schema.json. The mod's settings, its pages, the facts on its first page and its setup items. See [Settings and pages](sharing.md#settings-and-pages).

| Field | Type | What it is |
|---|---|---|
| `folders` | object of string by key | Folders a mod.folder:&lt;id> button opens, by id. |
| `live` | object | The facts on its first page. |
| `live.facts` | object of object by key | Facts by id, as status.fact sets them. |
| `live.facts.*.help` | string | Its tip. |
| `live.facts.*.label` | string | Its name on the card. |
| `live.facts.*.meter` | boolean | Drawn as a meter (0-100). |
| `live.facts.*.order` | integer | Its place on the card. |
| `live.facts.*.unit` | string | Shown after the value. |
| `pages` | object of object by key | The mod's pages, by id. |
| `pages.*.cards` | object of object by key | Its cards when it has no tabs, by id. |
| `pages.*.cards.*.help` | string | One line under the heading. |
| `pages.*.cards.*.order` | integer | Its place among the cards. |
| `pages.*.cards.*.rows` | list of object | Its widgets, in order. |
| `pages.*.cards.*.rows[].action` | string | button: trigger.try, feel.try, mod.demo, mod.folder:&lt;id>, mod.file:values. |
| `pages.*.cards.*.rows[].enabled_by` | string | A mod path or core.&lt;key> that must be on. |
| `pages.*.cards.*.rows[].fact` | string | meter and now: a live fact. |
| `pages.*.cards.*.rows[].groups` | list of string | levels: only these entry groups. |
| `pages.*.cards.*.rows[].help` | string | Overrides the setting's help. |
| `pages.*.cards.*.rows[].label` | string | Overrides the setting's label. |
| `pages.*.cards.*.rows[].pairs` | list of list of string | situations: keys shown as one row, R2's then L2's. |
| `pages.*.cards.*.rows[].path` | string | The setting it edits. |
| `pages.*.cards.*.rows[].preview` | string | color-grid: "pad" draws the controller. |
| `pages.*.cards.*.rows[].text` | string | note: its text. |
| `pages.*.cards.*.rows[].widget` | one of toggle, choice, select, slider, text, path, color, levels, rumble-table, situations, color-grid, chips, hex-chips, keyed-list, meter, now, button, note | The widget. |
| `pages.*.cards.*.title` | string | Its heading. |
| `pages.*.icon` | string | A name from the window's icon set. |
| `pages.*.lead` | string | The line under the page title. |
| `pages.*.order` | integer | Its place among the mod's pages. |
| `pages.*.tabs` | object of object by key | Its tabs, by id. |
| `pages.*.tabs.*.cards` | object of object by key | Its cards, by id. |
| `pages.*.tabs.*.cards.*.help` | string | One line under the heading. |
| `pages.*.tabs.*.cards.*.order` | integer | Its place among the cards. |
| `pages.*.tabs.*.cards.*.rows` | list of object | Its widgets, in order. |
| `pages.*.tabs.*.cards.*.rows[].action` | string | button: trigger.try, feel.try, mod.demo, mod.folder:&lt;id>, mod.file:values. |
| `pages.*.tabs.*.cards.*.rows[].enabled_by` | string | A mod path or core.&lt;key> that must be on. |
| `pages.*.tabs.*.cards.*.rows[].fact` | string | meter and now: a live fact. |
| `pages.*.tabs.*.cards.*.rows[].groups` | list of string | levels: only these entry groups. |
| `pages.*.tabs.*.cards.*.rows[].help` | string | Overrides the setting's help. |
| `pages.*.tabs.*.cards.*.rows[].label` | string | Overrides the setting's label. |
| `pages.*.tabs.*.cards.*.rows[].pairs` | list of list of string | situations: keys shown as one row, R2's then L2's. |
| `pages.*.tabs.*.cards.*.rows[].path` | string | The setting it edits. |
| `pages.*.tabs.*.cards.*.rows[].preview` | string | color-grid: "pad" draws the controller. |
| `pages.*.tabs.*.cards.*.rows[].text` | string | note: its text. |
| `pages.*.tabs.*.cards.*.rows[].widget` | one of toggle, choice, select, slider, text, path, color, levels, rumble-table, situations, color-grid, chips, hex-chips, keyed-list, meter, now, button, note | The widget. |
| `pages.*.tabs.*.cards.*.title` | string | Its heading. |
| `pages.*.tabs.*.lead` | string | The line under the tab bar. |
| `pages.*.tabs.*.order` | integer | Its place among the tabs. |
| `pages.*.tabs.*.title` | string | The tab's name. |
| `pages.*.title` | string | On the Mods page, in the mod's page links and in the page head. |
| `schema` | integer | The format: 1. |
| `settings` | object of a setting by key | Settings by id, in order. |
| `setup` | object of object by key | Setup checks it adds, by id. |
| `setup.*.fact` | string | The fact that holds its state. |
| `setup.*.help` | string | What to do when it fails. |
| `setup.*.label` | string | The check's name. |
| `setup.*.order` | integer | Its place in the list. |

### A setting (`setting`), used by `settings.*`, a map's `of` and `key` and a record's `fields.*`

One setting, or the type of a map's entries or a record's fields.

| Field | Type | What it is |
|---|---|---|
| `alias` | object of string by key | enum: file values shown as another choice. |
| `by_value` | object of string by key | enum: the help for each value. |
| `clamp` | one of window, file | window (default) or file. |
| `confirm` | object | A question before the window turns it on. |
| `confirm.ok` | string | The yes button. |
| `confirm.text` | string | What it means. |
| `confirm.title` | string | The question. |
| `default` | any | The value without a player's change. |
| `entries` | object of object by key | map and intset: words per key. |
| `entries.*.also` | list of string | Values allowed besides the type's. |
| `entries.*.card` | string | situations: the card title of this key. |
| `entries.*.group` | string | The group it is listed in. |
| `entries.*.help` | string | One line under it. |
| `entries.*.label` | string | The entry's name. |
| `entries.*.pattern` | string | colours: how it shows. |
| `entries.*.placeholder` | string | What an empty value means. |
| `entries.*.row` | string | situations: the row's name of a pair. |
| `entry_help` | string | map: the help of each entry. |
| `entry_label` | string | map: the label of an entry without words; {key} is its key. |
| `enum` | list of string | enum values without labels. |
| `fields` | object of a setting by key | record: its fields. |
| `groups` | object of object by key | Groups of entries. |
| `groups.*.help` | string | One line under the heading. |
| `groups.*.order` | integer | Its place among the groups. |
| `groups.*.title` | string | Its heading. |
| `help` | string | One line under it. |
| `key` | a setting | map: the keys' type when any key of it is allowed. |
| `label` | string | The setting's name. |
| `log` | boolean | A slider on a log scale. |
| `max` | number | The highest value. |
| `min` | number | The lowest value. |
| `more` | string | The longer help. |
| `needs` | one of haptics, rumble, triggers, lightbar, player_leds, mic, gyro, screen | The capability it needs. |
| `of` | a setting | map: the entries' type. |
| `off` | string | The line a page's header switch shows while off. |
| `off_help` | string | The help while a switch is off. |
| `options` | object of string by key | enum: value -> label, in order. |
| `placeholder` | string | A text field's placeholder. |
| `restart` | one of mod, session | How a change applies; none: at once. |
| `role` | one of feel_gain, feel_rumble, color_table, trigger_table, path, port | What it means to Bururu. |
| `row` | string | The row's own name where it differs from label. |
| `scale` | number | Shown as value x scale. |
| `step` | number | A slider's step. |
| `type` | one of bool, int, float, enum, path, rgb, hex, trigger, rumble, intset, map, record, string | The type. |
| `unit` | string | Shown after the value. |

## layout

A binary layout. How udp.listen decodes a game's packets. See [Binary layouts](layouts.md).

| Field | Type | What it is |
|---|---|---|
| `channels` | object of object by key | Vocabulary channels from fields. |
| `channels.*.abs` | boolean | The absolute value. |
| `channels.*.any` | boolean | True when any wheel holds. |
| `channels.*.clamp` | list of number | [lo, hi]. |
| `channels.*.div` | string | Divide by another field. |
| `channels.*.eq` | any | True when the value equals this. |
| `channels.*.from` | string | &lt;variant>.&lt;field>. |
| `channels.*.hypot` | list of string | The hypotenuse of fields. |
| `channels.*.ne` | any | True when the value differs from this. |
| `channels.*.offset` | number | Add. |
| `channels.*.scale` | number | Multiply. |
| `edges` | object of object by key | Edges evaluated per datagram. |
| `edges.*.above` | number | threshold: rises past this. |
| `edges.*.any` | boolean | Over wheels: any wheel. |
| `edges.*.below` | number | threshold: falls back under this. |
| `edges.*.kind` | one of change, rise, fall, threshold | The edge. |
| `edges.*.on` | string | A channel or &lt;variant>.&lt;field>. |
| `endian` | one of little, big | The byte order. |
| `enums` | object of object of string by key by key | Value names: enum name -> value -> name. |
| `gate` | object | No data while this does not hold (menus, pauses). |
| `gate.eq` | any | Holds when the field equals this. |
| `gate.field` | string | The field. |
| `gate.ne` | any | Holds when the field differs from this. |
| `gate.variant` | string | The variant. |
| `header` | object | The header. |
| `header.fields` | object of object by key | Its fields. |
| `header.fields.*.at` | integer | Its offset in bytes. |
| `header.fields.*.bits` | object of integer by key | bits32: named bits. |
| `header.fields.*.clamp` | list of number | [lo, hi]. |
| `header.fields.*.endian` | one of little, big | This field's byte order. |
| `header.fields.*.enum` | string | An enums map that names its values. |
| `header.fields.*.offset` | number | Add to the value. |
| `header.fields.*.order` | list of string | A per-wheel array's wheel order in the game: mapped to FL FR RL RR. |
| `header.fields.*.redact` | boolean | The recorder zeroes its bytes. |
| `header.fields.*.scale` | number | Multiply the value. |
| `header.fields.*.type` | string | u8 i8 u16 i16 u32 i32 u64 i64 f32 f64 bool8 bool32 char[N] wchar[N], arrays T[N], bits32. |
| `header.fields.*.unit` | string | Its unit. |
| `header.size` | integer | Its size in bytes. |
| `layout` | integer | The format: 1. |
| `unknown` | one of report, ignore | Datagrams no variant takes. |
| `variants` | object of object by key | Variants by name. |
| `variants.*.array` | object | An array of structs. |
| `variants.*.array.at` | integer | Its offset. |
| `variants.*.array.count` | any | The count, or a field that holds it. |
| `variants.*.array.pick` | string | A header field that picks the struct. |
| `variants.*.array.stride` | integer | Bytes per struct. |
| `variants.*.cases` | object of object by key | Cases by tag. |
| `variants.*.cases.*.fields` | object of object by key | Its fields. |
| `variants.*.cases.*.fields.*.at` | integer | Its offset in bytes. |
| `variants.*.cases.*.fields.*.bits` | object of integer by key | bits32: named bits. |
| `variants.*.cases.*.fields.*.clamp` | list of number | [lo, hi]. |
| `variants.*.cases.*.fields.*.endian` | one of little, big | This field's byte order. |
| `variants.*.cases.*.fields.*.enum` | string | An enums map that names its values. |
| `variants.*.cases.*.fields.*.offset` | number | Add to the value. |
| `variants.*.cases.*.fields.*.order` | list of string | A per-wheel array's wheel order in the game: mapped to FL FR RL RR. |
| `variants.*.cases.*.fields.*.redact` | boolean | The recorder zeroes its bytes. |
| `variants.*.cases.*.fields.*.scale` | number | Multiply the value. |
| `variants.*.cases.*.fields.*.type` | string | u8 i8 u16 i16 u32 i32 u64 i64 f32 f64 bool8 bool32 char[N] wchar[N], arrays T[N], bits32. |
| `variants.*.cases.*.fields.*.unit` | string | Its unit. |
| `variants.*.emit` | string | event: each packet is an input. |
| `variants.*.fields` | object of object by key | Its fields. |
| `variants.*.fields.*.at` | integer | Its offset in bytes. |
| `variants.*.fields.*.bits` | object of integer by key | bits32: named bits. |
| `variants.*.fields.*.clamp` | list of number | [lo, hi]. |
| `variants.*.fields.*.endian` | one of little, big | This field's byte order. |
| `variants.*.fields.*.enum` | string | An enums map that names its values. |
| `variants.*.fields.*.offset` | number | Add to the value. |
| `variants.*.fields.*.order` | list of string | A per-wheel array's wheel order in the game: mapped to FL FR RL RR. |
| `variants.*.fields.*.redact` | boolean | The recorder zeroes its bytes. |
| `variants.*.fields.*.scale` | number | Multiply the value. |
| `variants.*.fields.*.type` | string | u8 i8 u16 i16 u32 i32 u64 i64 f32 f64 bool8 bool32 char[N] wchar[N], arrays T[N], bits32. |
| `variants.*.fields.*.unit` | string | Its unit. |
| `variants.*.min_len` | integer | The shortest datagram it takes. |
| `variants.*.tag` | object | A tag inside the packet that chooses a case. |
| `variants.*.when` | object | Header field values or a datagram length that choose it. |

## screen

A screen layout, or a number font beside it. What the screen sensor reads, where, and what its changes mean. See [Screen layouts](sensors.md#screen-layouts).

| Field | Type | What it is |
|---|---|---|
| `capped` | boolean | A font file: its numbers never pass 100%. |
| `colours` | object of object by key | The colours to look for, by name. |
| `colours.*.aliases` | list of string | Other names the player's colour setting may use. |
| `colours.*.default` | string | The game's standard colour, such as "#29c8cf". |
| `colours.*.label` | string | How the log names it; none: left out of the log. |
| `colours.*.match` | object | How close a pixel must be. |
| `colours.*.match.chroma_tol` | number | Pale colours (white, grey): how far the colour may be off, brightness left out. |
| `colours.*.match.hue_width` | number | Degrees either side of the colour's hue. |
| `colours.*.match.min_sat` | number | Saturation at least this part of the colour's. |
| `colours.*.match.min_val` | number | Brightness at least this part of the colour's. |
| `colours.*.off_near` | object | The colour is off while it is this close to another. |
| `colours.*.off_near.colour` | string | The other colour. |
| `colours.*.off_near.distance` | number | How close. |
| `font` | integer | A font file: the format, 1. |
| `fonts` | object of string by key | The number fonts by name: font files in the mod, such as screen/digits.json. |
| `glyphs` | object of list of number by key | A font file: the digits 0 to 9, and "%" when the game shows one, each 240 numbers (a 12 x 20 cell, row by row). |
| `learn` | object | How the colours are found on screen: by reading two numbers. |
| `learn.also` | string | A part whose colour is close to pair's, found by reading it. |
| `learn.dx` | list of number | Where pair's centre is right of part's, in glyph heights, [low, high]. |
| `learn.dy` | list of number | Where pair's centre is below part's, in glyph heights, [low, high]. |
| `learn.near` | number | How far also's colour may be from pair's. |
| `learn.pair` | string | A number part near it. |
| `learn.part` | string | A number part with an area. |
| `matrix` | object | The elements of the game's colour matrix file: the section, and its three rows (red, green, blue in). |
| `matrix.rows` | list of string | The three rows. |
| `matrix.section` | string | The section. |
| `name` | string | How the log names what is read; default: Screen. |
| `parts` | object of object by key | What to read and where, by name, in the order changes are taken. |
| `parts.*.alert` | string | lists: the colour of an entry that cannot be used now. |
| `parts.*.alone_min` | number | pair: a main number without its second must score this much. |
| `parts.*.anchor` | one of top-left, top, top-right, left, centre, right, bottom-left, bottom, bottom-right | The area is in the widest 16:9 box that fits the window, with the box at this place of it, so the area fits other sizes and shapes of window. |
| `parts.*.area` | list of number | Where it is: x0 y0 x1 y1 as parts of the window, 0 to 1 (lists: screen heights from the centre). |
| `parts.*.at` | list of number | A number next to another: where it usually is in its box. |
| `parts.*.box` | list of number | Next to a number: x0 y0 x1 y1 in that number's glyph heights from its centre. |
| `parts.*.colour` | string | The colour to look for, from colours (bar: its fill's). |
| `parts.*.column` | number | zone: its width, in glyph heights. |
| `parts.*.dim` | object | bars: an unlit block's colour. |
| `parts.*.dim.hue` | number | Degrees either side of the hue. |
| `parts.*.dim.sat` | number | The least saturation. |
| `parts.*.dim.val` | number | The least brightness. |
| `parts.*.empty` | string | bar: the colour of its empty part, from colours; none: not looked for. |
| `parts.*.events` | object | What its changes mean: each comes as the input &lt;sensor>:&lt;kind>. |
| `parts.*.events.break` | an event | pair: the main number went and the second shows alone. |
| `parts.*.events.busy` | string | lists: the input when an entry's status line shows. |
| `parts.*.events.done` | string | lists: the input when it is gone again. |
| `parts.*.events.drop` | an event | number, pair, bar: the value dropped; x is the drop (a bar's in percent of the bar). |
| `parts.*.events.flash` | an event | number, pair: a flash came up; x is its level. |
| `parts.*.events.hide` | an event | icon: the picture is gone; ev.picture names it. |
| `parts.*.events.notch` | an event | number: the value rose past from, from + step, ... up to to; x is the notch. |
| `parts.*.events.rise` | an event | number, bar: the value rose. |
| `parts.*.events.second_drop` | an event | pair: the second number dropped; x is the drop. |
| `parts.*.events.show` | an event | icon: a picture shows, or another takes its place; ev.picture names it. |
| `parts.*.every` | integer | Read every this many reads; default 1. |
| `parts.*.filter` | object | number, pair: how reads become a steady value. |
| `parts.*.filter.max` | integer | The highest value shown. A plain number: optional, it sets how many digits it has. |
| `parts.*.filter.min_score` | number | The least score a read needs. |
| `parts.*.filter.near` | integer | A drop by up to this much is taken at once. |
| `parts.*.filter.need` | integer | Reads in a row that take any other value. |
| `parts.*.flash` | object | number, pair: the area of a colour above it, such as hit flashes. |
| `parts.*.flash.box` | list of number | x0 y0 x1 y1 in the number's glyph heights from its centre. |
| `parts.*.flash.colour` | string | The flash's colour. |
| `parts.*.flash.level` | string | A name the flash's level shows under; none: kept inside. |
| `parts.*.flash.spread` | number | Glyph heights from the centre that read as fully left or right. |
| `parts.*.follow` | object | number, pair: the capture window kept around it while it reads well. |
| `parts.*.follow.box` | list of number | Around the number, in its glyph heights. |
| `parts.*.follow.min_score` | number | The least score that keeps the window. |
| `parts.*.follow.misses` | integer | pair: reads without it before the window is dropped. |
| `parts.*.follow.search` | list of number | Inside the window, where the number is looked for; none: all of it. |
| `parts.*.font` | string | number, pair: the font, from fonts. |
| `parts.*.forget_ms` | integer | pair: not seen this long, its numbers start over. |
| `parts.*.from` | one of left, right, top, bottom | bar: the end a full bar starts from; default left. |
| `parts.*.full` | number | bars: the least full height, in glyph heights. |
| `parts.*.gap` | list of number | bars: from one bar's centre to the next, in glyph heights, [low, high]. |
| `parts.*.gap_ms` | integer | zone: read at most this often. |
| `parts.*.glyph` | list of number | lists: a letter's height as parts of the screen height, [low, high]. |
| `parts.*.height` | list of number | bars: a bar's lit block in glyph heights, [low, high]. |
| `parts.*.hue` | list of number | zone: the column's hue range, in degrees. |
| `parts.*.icon` | object | number: an icon of another colour left of it, which tells it apart from other numbers. |
| `parts.*.icon.colour` | string | The icon's colour. |
| `parts.*.icon.fill` | list of number | The part of its box it covers, [low, high]. |
| `parts.*.icon.height` | list of number | Its height as parts of the screen height, [low, high]. |
| `parts.*.icon.ratio` | list of number | Its width / height, [low, high]. |
| `parts.*.icon.text` | list of number | Where the number is: x0 y0 x1 y1 in icon heights from its top right. |
| `parts.*.icon.text_size` | list of number | The number's glyph height in icon heights, [low, high]. |
| `parts.*.icon.text_start` | number | The number starts within this many icon heights. |
| `parts.*.icon_width` | number | lists: an entry's icon and the gap after it. |
| `parts.*.images` | object of string by key | icon: its pictures by name: PNG files in the mod, such as screen/horse.png. |
| `parts.*.kind` | one of number, pair, bars, zone, lists, bar, icon | What it is: a number ("NN%" or plain), a pair of numbers, bars or a zone next to a number, lists of names, a bar without a number, or an icon. |
| `parts.*.letter` | number | lists: the width of a letter not listed. |
| `parts.*.letter_gap` | number | lists: the gap between letters. |
| `parts.*.letters` | object of number by key | lists: letter widths in glyph heights, by letter. |
| `parts.*.lit` | object | bars: a lit block's colour. |
| `parts.*.lit.hue` | number | Degrees either side of the hue. |
| `parts.*.lit.sat` | number | The least saturation. |
| `parts.*.lit.val` | number | The least brightness. |
| `parts.*.marker` | list of number | zone: where the marker is right of it, in glyph heights. |
| `parts.*.match` | object of object by key | lists: how close a pixel must be for full, dim, alert and alert_dim. |
| `parts.*.match.*.chroma_tol` | number | Pale colours (white, grey): how far the colour may be off, brightness left out. |
| `parts.*.match.*.hue_width` | number | Degrees either side of the colour's hue. |
| `parts.*.match.*.min_sat` | number | Saturation at least this part of the colour's. |
| `parts.*.match.*.min_val` | number | Brightness at least this part of the colour's. |
| `parts.*.min_score` | number | number, pair, zone: the least a number's match may score to count. |
| `parts.*.names` | list of string | bars: the bars' names, left to right. |
| `parts.*.near` | string | number, bars, zone: the number part with an area it sits next to. |
| `parts.*.offset` | integer | With every: start at this read. |
| `parts.*.percent` | boolean | number: false reads a plain number without "%"; default true. |
| `parts.*.sat` | list of number | zone: its saturation range. |
| `parts.*.scan_ms` | integer | pair: while lost, look for it this often (only while reading fast). |
| `parts.*.second` | object | pair: the second number. |
| `parts.*.second.alone` | list of number | The least score of a second alone: while searching, and where it was. |
| `parts.*.second.at` | list of number | Where it usually is in the box. |
| `parts.*.second.box` | list of number | x0 y0 x1 y1 in the main number's glyph heights from its centre. |
| `parts.*.second.colour` | string | Its colour. |
| `parts.*.second.filter` | object | How its reads become a steady value. |
| `parts.*.second.filter.max` | integer | The highest value shown. A plain number: optional, it sets how many digits it has. |
| `parts.*.second.filter.min_score` | number | The least score a read needs. |
| `parts.*.second.filter.near` | integer | A drop by up to this much is taken at once. |
| `parts.*.second.filter.need` | integer | Reads in a row that take any other value. |
| `parts.*.second.min_score` | number | The least its match may score. |
| `parts.*.second.name` | string | Its name, its own. |
| `parts.*.second.size` | number | Its glyph height is at least this part of the main number's. |
| `parts.*.segment_gap` | number | bar: the widest gap in its fill, as a part of its length, such as the lines between segments; default 0.03. |
| `parts.*.sides` | list of string | lists: the left and the right list's names. |
| `parts.*.size` | number | icon: the icon's height as a part of the window height. |
| `parts.*.smooth` | integer | bar: its value is the middle one of this many reads; default 3. icon: reads in a row that change what shows; default 2. |
| `parts.*.span` | list of number | zone: where it may be, as parts of the box's width. |
| `parts.*.tolerance` | number | icon: how much the colours may differ on average, 0 to 1; default 0.12. |
| `parts.*.trail` | number | bar: fill dimmer than this part of the fill's start is the trail of a value just lost; 0: off; default 0.7. |
| `parts.*.val` | number | zone: its least brightness. |
| `parts.*.width` | list of number | bars: a bar's width in glyph heights, [low, high]. |
| `parts.*.word_gap` | number | lists: the gap between words. |
| `screen` | integer | The layout format: 1. |

### A part's event (`event`), used by `parts.*.events`

A change a part reports, as the input &lt;sensor>:&lt;kind> with a strength: min(1, base + (x + add) / div).

| Field | Type | What it is |
|---|---|---|
| `add` | number | Added to x. |
| `base` | number | The strength before x. |
| `below` | number | break: the main value was at most this. |
| `damage` | boolean | drop: the drop goes in ev.damage. |
| `div` | number | Divides x + add; default 1. |
| `from` | integer | notch: the first value. |
| `gap_ms` | integer | At least this long after the part's last hit; 0: no wait. |
| `kind` | string | The input's name. |
| `pan_ms` | integer | drop: the pan of a flash seen this lately. |
| `reads` | integer | break: reads with the second alone; second_drop: after this many. |
| `step` | integer | notch: from one value to the next. |
| `to` | integer | notch: the last value. |
| `unread` | boolean | flash: only while the number has no steady value. |

## bindings

A bindings layout: how a game writes its control preset, an XML file. The bindings sensor reads the preset with it. See [bindings](sensors.md#bindings).

| Field | Type | What it is |
|---|---|---|
| `actions` | object of string by key | The actions the mod reacts to, in order: the mod's name -> the preset's element. |
| `axes` | object of string by key | The axes it reads: the mod's name -> the preset's element. |
| `axis` | string | An axis's element that holds a stick. |
| `bindings` | integer | The format: 1. |
| `dead_zone` | string | An axis's element that holds its dead zone. |
| `dead_zones` | list of string | Names from axes whose dead zones are read. |
| `device` | string | The attribute that names the device. |
| `key` | string | The attribute that names the button, key or stick. |
| `keyboard` | object | How the preset names the keyboard. |
| `keyboard.devices` | list of string | Its device names. |
| `keyboard.prefix` | string | The start of its key names, such as Key_ in Key_Space; the rest is Bururu's name for the key: A, F9, Space. |
| `keys` | list of string | An action's elements that hold a button or key, such as Primary and Secondary. |
| `look` | object | Head look: an action that moves the view with the mouse while held, or turns it on and off. |
| `look.action` | string | The action's element. |
| `look.mouse` | string | The setting element whose value 1 lets the mouse look. |
| `look.toggle` | string | Its element whose value 1 makes it a toggle. |
| `modifier` | string | The element inside them for a button or key held with it. |
| `mouse` | object | How the mouse turns the view, per axis: [X, Y]. |
| `mouse.decay` | list of string | The setting elements whose value 0 keeps a turn where the mouse left it. |
| `mouse.turn` | list of string | The setting elements of what the mouse turns; any value means it turns. |
| `pads` | list of object | How the preset names each kind of controller. |
| `pads[].buttons` | object of one of square, cross, circle, triangle, l1, r1, l2, r2, create, options, l3, r3, ps, touchpad, up, down, left, right by key | The preset's key names -> the controller's buttons. |
| `pads[].devices` | list of string | The device names the preset gives it, or its USB ids. |
| `pads[].sticks` | object of one of lx, ly, rx, ry by key | The preset's key names -> the controller's stick axes. |
| `preset` | string | The root element's attribute that names the preset. |
| `turn` | list of string | Names from axes that turn: a stick bound to one is a turn stick. |
| `value` | string | The attribute that holds a setting's value. |

## migration

A settings and feel migration, applied once per player. See [Updates and migrations](sharing.md#updates-and-migrations).

| Field | Type | What it is |
|---|---|---|
| `aliases` | object | Old names other mods may still use. |
| `aliases.hooks` | object of string by key | Hook points, old -> new (name@version). |
| `aliases.rules` | object of string by key | Rule ids, old -> new. |
| `applies_below` | string | The version that renamed things; due from it on. |
| `drop_settings` | list of string | Dotted paths removed. |
| `id` | string | The file's name: 002-rename-brake-feel. |
| `map_values` | object of object of any by key by key | A dotted path ("\*" for every key) -> old value -> new value. |
| `migration` | integer | The format: 1. |
| `rename_feels` | object of string by key | Feel names, old -> new. |
| `rename_settings` | object of string by key | Dotted setting paths, old -> new. |

## vocab

A vocabulary. Channel names with units and ranges, kept in a library mod. See [Vocabularies](sharing.md#vocabularies).

| Field | Type | What it is |
|---|---|---|
| `channels` | object of object by key | Channels by full name. |
| `channels.*.enum` | list of string | The strings it may hold. |
| `channels.*.note` | string | What the values mean. |
| `channels.*.range` | list of number | The values Bururu accepts, [low, high]. |
| `channels.*.type` | one of bool, int | bool or int; none: a number. |
| `channels.*.unit` | string | Its unit. |
| `channels.*.wheels` | boolean | One value per wheel. |
| `version` | integer | The version after the @. |
| `vocab` | string | Its name: racing in racing@1. |
| `wheels` | list of string | The order of per-wheel arrays: FL FR RL RR. |

## soundprint

A sound print: where a clip's spectrogram has its peaks. It cannot be played back.

| Field | Type | What it is |
|---|---|---|
| `level_db` | number | The clip's loudest 8 ms, dBFS. |
| `made_with` | string | The Bururu that made it. |
| `ms` | integer | The clip's length, at most 2000. |
| `name` | string | The clip's name. |
| `other_starts` | list of integer | For each frame from the start: how many more peaks started there than the print keeps. |
| `peaks` | list of list of integer | The peaks, in frame order. |
| `soundprint` | integer | The format: 1. |

