# Sharing a mod

Back to the [modding guide](index.md).

What a mod needs once other people use it: the manifest, the pages players see, permissions, add-ons and updates. Every field is in the [file reference](files.md). Releasing a mod and listing it: [Publishing a mod](publishing.md).

## The manifest

```jsonc
{
  "manifest": 1,
  "id": "gravel-rally",
  "name": "Gravel Rally",
  "version": "1.0.0",
  "api": "0.12",
  "kind": "game",
  "authors": ["Your name"],
  "license": "MIT",
  "summary": "Gear changes, kerbs and the brake trigger from the game's telemetry.",
  "homepage": "https://github.com/you/bururu-gravel-rally"
}
```

- `id`: a lower-case letter, then 1 to 39 lower-case letters, digits or `-`. It never changes. A mod folder has the same name; a packed mod is `<id>-<version>.brr`.
- `version`: `major.minor.patch`. Raise it with each release.
- `api`: the mod API the mod is made for. This Bururu speaks `"0.12"` and still loads mods made for `"0.10"` and `"0.11"`, because 0.11 and 0.12 only added to them. 0.12 added the `macos` entries, see [Games on macOS](#games-on-macos), and the `game.output` sensor; a mod without them can keep `"0.11"`, so Bururus from before 0.12 run it too. A mod made for another version is refused with a message that says so. From 1.0 on, a mod made for 1.2 runs on every 1.x from 1.2 up.
- `name`, `summary`, `authors`, `license` and `homepage` show on the **Mods** page.
- `icon.png` next to the manifest is the mod's icon: a square PNG of 64 to 256 pixels, at most 32 KB.
- `os`: where the mod works, such as `["windows"]`.

A game mod's `games` entry names its game: `name`, `short_name` (Bururu's messages use it), the program under `process`, the window, and the controller the game sees:

| `virtual_pad` | What the game sees on Automatic |
|---|---|
| `game` | Bururu's virtual DualShock 4, on the cable and on Bluetooth |
| `launcher_or_game` | the same, from when the game's launcher starts |
| `never` | the DualSense itself; on Bluetooth Bururu leaves it alone, and the cable still works |
| `shared` | for a program that never reads a controller: Bururu changes nothing about the controller for it, and the player gets no choice. On Bluetooth with the drivers the DualSense stays as it is with no game (hidden, and games see the virtual DualSense); without them Bururu opens the DualSense on Bluetooth while it runs |

When the mod list has a profile for the game, Automatic follows the profile. The player can pick another controller for the game in Bururu.

A game that drives a DualSense itself says `"native_dualsense": true`, and in `native_hint` where its own vibration and trigger effects are switched off. When the game reads the DualSense itself (Direct), Bururu leaves it the controller while it runs. When it reads the DualSense while Bururu drives it on the cable (without the drivers for the virtual controller), the mod's card asks the player to switch them off, or to pick Direct.

### Finding the game

The key of each `games` entry is the game's id, written like a mod id: a lower-case letter, then 1 to 39 lower-case letters, digits or `-`, such as `gravel-rally`. Bururu's game scan finds installed games on the player's PC by these ids, from the records of the mod list. Use the list's id for your game when it has one.

For a game the list does not know, the manifest says it itself:

```jsonc
"games": {
  "gravel-rally": {
    "name": "Gravel Rally",
    "process": { "windows": { "game": ["GravelRally.exe"] } },
    "stores": { "steam": [1234560], "epic": [{ "app": "GravelRally" }] },
    "markers": ["GravelRally.exe", "Data/*/base.pak"],
    "virtual_pad": "game"
  }
}
```

- `stores`: the game's ids in each store: `steam`, `gog` and `ubisoft` (numbers), `epic` (`app`), `xbox` (`family`, `store_id`), `battlenet` and `ea`.
- `markers`: files that prove an install folder, relative to it, with `*` for one part.
- When the game runs, Bururu also learns its folder from the running program.

A declared path can name the game's install folder: `{GameDir}`, or `{GameDir:<game id>}` in a mod with several games. When no install is found, the path's sensor is "unavailable here" and the mod still loads.

### Games on Linux

Most game mods need nothing extra for Linux. A Windows game runs there through Proton or Wine, and Bururu finds it with what the mod says for Windows:

- With no `linux` list under `process`, Bururu looks for the Windows program names. The game keeps them under Proton.
- `steam_app` gives the game's window, which Proton names `steam_app_<id>`, and its Proton prefix, the folder that holds the game's Windows folders. The Windows folders in a path, such as `{SavedGames}`, are the ones in that prefix.
- For a game from another launcher, such as Heroic or Lutris, Bururu learns its Wine prefix when it sees the game run.
- Bururu sees the game's window in front, reads keys for the bindings sensor and reads the screen only when the game runs in an X11 window. Proton gives every game one, unless the player turns on the Wayland mode some Proton builds have.
- Leave `os` out of the manifest, or list both systems: `["windows", "linux"]`. With `["windows"]` the mod says "Windows only" on Linux and does not run.

A native Linux game gets its own entries: its program in `process` as `"linux": { "game": ["gravelrally"] }`, and its folders in the `linux` form of each path. There `{LocalAppData}` is `~/.local/share`, `{RoamingAppData}` is `~/.config`, `{Documents}` is the Documents folder, and `{SavedGames}` and `{Home}` are the home folder:

```jsonc
"paths": {
  "logs": { "path": "{Documents}/Gravel Rally/logs", "linux": "{LocalAppData}/gravel-rally/logs", "label": "Game logs" }
}
```

### Games on macOS

Bururu finds a game on a Mac the same way. Every `macos` entry below needs `"api": "0.12"`, so a Bururu from before macOS says the mod needs a newer Bururu.

- A native Mac game names its program in `process` as `"macos": { "game": ["GravelRally"] }`: the file in the app's `Contents/MacOS` folder.
- With no `macos` list under `process`, Bururu looks for the Windows program names. That finds a Windows game that runs in a CrossOver bottle.
- Steam keeps its library in `~/Library/Application Support/Steam`, so `{SteamApp:<id>}` and the game scan find Steam's Mac games. A Mac game's `{GameDir}` is the folder that holds its `.app`.
- For a Windows game in a bottle, Bururu learns the bottle when it sees the game run. The Windows folders in a path, such as `{SavedGames}`, are then the ones in that bottle. `proton` plays no part on a Mac.
- A native Mac game's folders go in the `macos` form of each path. There `{LocalAppData}` and `{RoamingAppData}` are `~/Library/Application Support`, `{Documents}` is the Documents folder, and `{SavedGames}` and `{Home}` are the home folder.
- `window` takes the app's bundle id on macOS, `"macos": { "bundle_id": "com.example.gravelrally" }`, for the screen sensor once it runs there. `osascript -e 'id of app "Gravel Rally"'` prints it.
- The screen, bindings and memory sensors do not run on macOS yet. A mod that has them runs on its other sensors.
- On macOS a DualSense on its cable has no wave haptics: `t.caps.voices` is false and `voice_rate` 0, so the feels play on its motors. On Bluetooth they play as native haptics, as elsewhere.
- Leave `os` out of the manifest, or list `macos` with the others. With `["windows", "linux"]` the mod says "made for Windows and Linux" on a Mac and does not run.
- A Mac App Store game keeps its files in a folder macOS guards, and macOS may ask the player before Bururu reads it. A Steam build is the easier start.

```jsonc
"process": { "windows": { "game": ["GravelRally.exe"] }, "macos": { "game": ["GravelRally"] } },
"paths": {
  "logs": { "path": "{Documents}/Gravel Rally/logs", "macos": "{Home}/Library/Application Support/Gravel Rally/logs", "label": "Game logs" }
}
```

## Settings and pages

`settings.schema.json` declares a mod's settings and the pages the window draws for them. A mod has no window code: the window draws its pages with the same controls as Bururu's own.

```jsonc
{
  "schema": 1,
  "settings": {
    "gear_feel": { "type": "enum", "default": "clunk", "label": "Gear feel",
                   "options": { "clunk": "Clunk", "click": "Click", "off": "Off" } },
    "haptics_gain": { "type": "map", "role": "feel_gain", "needs": "haptics",
                      "of": { "type": "float", "min": 0, "max": 2, "step": 0.05, "unit": "%", "scale": 100 },
                      "default": { "tap": 1, "swell": 1 },
                      "label": "Effect levels", "help": "100% is as made, 0% turns it off.",
                      "entries": { "tap": { "label": "Tap" }, "swell": { "label": "Swell" } } }
  },
  "pages": {
    "effects": { "title": "Effects", "icon": "waves", "order": 1, "cards": {
      "gears":  { "title": "Gears", "order": 1, "rows": [ { "widget": "choice", "path": "gear_feel" } ] },
      "levels": { "title": "Effect levels", "order": 2, "rows": [ { "widget": "levels", "path": "haptics_gain" } ] } } }
  },
  "live":  { "facts": { "mode": { "label": "Mode", "order": 1 } } }
}
```

- **Types**: `bool`, `int`, `float`, `enum`, `string`, `path`, `rgb`, `hex`, `trigger`, `rumble`, `intset`, `map` and `record`. Every setting has a `default` and a `label`.
- **Roles** tell Bururu what a setting is: `feel_gain` (the effect levels), `feel_rumble` (each feel on the motors), `color_table` (the lightbar colours `lights.json` names), `trigger_table` (the trigger effects), `path` (a folder a declared path follows), `port` (a UDP port).
- **Widgets**: `toggle`, `choice`, `select`, `slider`, `text`, `path`, `color`, `levels`, `rumble-table`, `situations` (trigger effects, with **Try**), `color-grid`, `chips`, `hex-chips`, `keyed-list`, `meter`, `now`, `button` and `note`.
- `help` is a line under a setting; `needs` greys a row out where the controller cannot do it; `restart` restarts the mod (`mod`) or Bururu's loop (`session`) when the setting changes.
- `live.facts` are what the mod's first page shows, set from Lua with `status.fact(id, value)`. `setup` adds checks to the **Controller** page, set from Lua with `setup.item(id, state, text)`.

The player's changes are kept in `mod-settings\<id>.json`, and only what differs from the defaults. So when an update improves a default, every player who never changed it gets it. Lua reads the values as `t.settings`, rules as `settings`.

## Packing

```
.\brr mod pack mods\my-data-mod --out .
```

`mod pack` checks the mod and writes `<id>-<version>.brr`, a zip, without the editor support. The same files always give the same zip, and it prints its SHA-256.

- Allowed files: `.json`, `.jsonl`, `.lua`, `.md`, `.txt`, `.png`, `.wav`, `.glyphs`, and licence files. Anything else refuses the mod.
- Limits: 64 MB for the zip, 16 MB for a file, 4 MB for a JSON file, 4,000 files.
- `.git`, `.github`, `.gitignore` and `.gitattributes` are skipped, so a clone of the mod's repo loads as a folder.
- REAPER projects (`.rpp`) are skipped too, so a mod's repo can keep the project its feels come from.

A player installs the pack with **Install from file...**, from the mod list once it is listed, or by putting it into the `mods` folder. Bururu reads it in place.

When there are two copies of one mod, a folder beats a `.brr`, and both beat a built-in mod. Two folders or two packs with one id: neither loads, and the **Mods** page names both.

## Permissions

A mod reads only what it declares, and the player agrees first:

```jsonc
"paths": {
  "logs": { "path": "{Documents}/Gravel Rally/logs", "label": "Game logs" }
},
"permissions": { "files": ["logs"], "screen": false, "keyboard": false, "udp": [20777] }
```

| Permission | Allows |
|---|---|
| `files` | reading the declared paths it names (never writing) |
| `screen` | the screen sensor |
| `keyboard` | the bindings sensor, which sees the keys bound to the game's actions while the game is in front |
| `memory` | the memory sensor, which reads the game's memory on Windows, for a game the mod list marks as offline with no anti-cheat |
| `sound` | the sound sensor, which listens to the game's own sound while it runs, on Windows 11, and keeps nothing |
| `udp` | listening on these UDP ports |
| `udp_forward` | passing packets on to these addresses off this PC |

The install dialog lists every folder (as found on the player's PC), the screen, the keyboard, the game's memory, the game's sound, each port and each forward address, for the mod and each mod it needs, before anything downloads. An update that asks for more asks again. A mod put into the `mods` folder by hand asks with **Turn on**. `{ModData}` and the mod's own files need no permission.

A path must start with a path token and name a folder below a Windows folder, never the folder itself. A game with kernel-level anti-cheat says so (`"anti_cheat": { "kind": "kernel" }`), and Bururu then keeps that mod's screen and keyboard sensors off. Bururu reads a game's memory only for a game the mod list marks as offline with no anti-cheat, see [memory](sensors.md#memory). It listens to a game's sound only in a program the mod list names for the game, see [sound](sensors.md#sound).

## Dependencies and load order

```jsonc
"dependencies": {
  "rally-feels": { "kind": "required", "version": ">=0.9.0 <2.0.0" },
  "rally-sounds": { "kind": "optional" },
  "old-rally":   { "kind": "incompatible", "why": "changes the same feels" }
}
```

- `required` (the default): without it, or out of the version range, the mod does not load, and the message says what to install. `optional` only orders the load. `incompatible`: when both are on, this one does not load.
- A dependency loads first. `load_after` and `load_before` order without depending. An add-on loads after the mods it changes. Otherwise the player's order on the **Mods** page decides.

## Add-ons

An add-on (`"kind": "addon"`) changes other mods by name, without copying them:

```jsonc
"kind": "addon",
"for": ["gravel-rally"],
"dependencies": { "gravel-rally": { "kind": "required", "version": ">=1.0.0 <2.0.0" } }
```

Its files under `overrides/<mod id>/` change that mod:

| What | Replace | Patch |
|---|---|---|
| a feel | `feels/<name>.feel.json` | `feels/<name>.feel.patch.json`, voices by id |
| rules and layers | the same id in `rules/<any>.rules.json` | `rules/<any>.patch.json`, by rule id; `null` removes |
| light rows and overlays | the same id in `lights.json` | `lights.patch.json`; `null` removes |
| a sensor | | `sensors.patch.json`: only the fields its `patchable` lists |
| setting defaults | | `settings.defaults.json`; players who changed a setting keep their value |
| pages | | `settings.pages.json`, by page, tab and card |
| a data table | `data/<name>.json` | `data/<name>.patch.json` |
| a layout | `layouts/<name>.layout.json` | a patch by variant and field |
| a Lua file | `lua/<file>.lua` | |

A patch changes the fields it names, and `null` removes. An add-on may also bring its own feels, rules, rows and Lua. Files under `final/<mod id>/` apply after every other add-on's changes. Prefer patches and [hook points](#hook-points) to replacing whole files: they keep working when the mod updates.

`mod check --conflicts`, and the **Conflicts** tab of the **Mods** page, list who changes what and who wins. Real conflicts are marked: two mods on one port, two game mods for one program, a whole Lua file replaced, a patch on something that does not exist.

## Hook points

A game mod lets add-ons change what its Lua is about to do at named points. It declares them in its manifest:

```jsonc
"hooks": { "gear@1": { "doc": "Before the gear clunk: change level, or skip it." } }
```

and runs them with a table of plain values:

```lua
local p = hook.run("gear@1", { level = 1, skip = false })
if not p.skip then feel.play("gear_clunk", { scale = p.level }) end
```

An add-on attaches a handler:

```lua
hook.add("gravel-rally/gear@1", "quiet-gears/gear", function(p)
  p.level = p.level * 0.6
end)
```

Handlers run in load order, each on a copy of the table; a handler that fails is switched off. With no handlers, `hook.run` gives the table back unchanged. When a point's meaning changes, its version goes up (`gear@2`). `mod check` fails an add-on that names a point the game mod does not declare.

Mods that run together can also share functions: `bururu.export("name", fn)` in one, `bururu.import("<mod id>", "name")` in another.

## Vocabularies

A vocabulary is a list of channel names with units and ranges, kept in a library, so one feel pack works for every game that publishes those names. Bururu has none built in; anyone can publish one in a library. Say a library `racing` holds the vocabulary `racing@1`:

- A game mod that publishes them says `"provides": ["racing@1"]` and depends on `racing`. It publishes channels from its layout, from rules (`{ "channel": { "racing.boost": "..." } }`) or from Lua (`channels.set("racing.boost", v)`). Bururu checks each value's name, type and range.
- A feel pack is an add-on with `"for": ["racing@1"]`. It comes with whichever of those games drives, and reads only `ch['racing.*']`.
- Wheel channels are lists: front left, front right, rear left, rear right.

A library keeps a vocabulary in `vocab/<name>.vocab.json`:

```jsonc
{ "vocab": "flight", "version": 1, "channels": {
  "flight.altitude": { "unit": "m" },
  "flight.gear_down": { "type": "bool" },
  "flight.throttle": { "range": [0, 1] } } }
```

## Updates and migrations

When an update renames a setting or a feel, a migration moves the players' values along:

```jsonc
{ "migration": 1, "id": "002-rename-clunk", "applies_below": "1.2.0",
  "rename_settings": { "haptics_gain.clunk": "haptics_gain.gear_clunk" },
  "rename_feels":    { "clunk": "gear_clunk" },
  "map_values":      { "gear_feel": { "thud": "clunk" } },
  "drop_settings":   ["old_switch"] }
```

Put it in `migrations/<number>-<name>.json` in the version that renames things. Bururu applies it once per player: it rewrites the player's values file and keeps the old one as `.bak`.
