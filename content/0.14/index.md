# Making mods

A mod tells Bururu about one game: which program it is, what to read while it runs, and what each moment plays on the controller. A mod is a folder of text files. You need Bururu, a DualSense on its USB cable, and a text editor. There is nothing to compile.

## Quick start

This takes about 10 minutes. You make a mod from the `data` template, feel it, change it, check it, test it and pack it. The template's "game" is Notepad, and a small JSON file you edit in Notepad stands in for a file a real game writes.

On Linux, run the commands as [On Linux](#on-linux) shows, and use GNOME Text Editor or gedit where the steps say Notepad: the template looks for those two there.

### Make it

Open PowerShell in Bururu's folder: in Explorer, right-click an empty spot in the folder and choose **Open in Terminal**. Then:

```
.\brr mod new my-data-mod --template data
```

This writes the mod into `mods\my-data-mod`, and a first `state.json` into `mod-data\my-data-mod`, the mod's own data folder.

### Feel it

1. Start Bururu. Its **Mods** page lists **Notepad**.
2. Open `mod-data\my-data-mod\state.json` in Notepad. While Notepad runs, the mod drives the controller: the lightbar turns blue.
3. Set `"Flags"` to `1` and save. You feel a tap, and the lightbar blinks red.
4. Set it back to `0` and save: a tap on the left.
5. Set it to `2` and save: a white flash, and both triggers resist.
6. Set `"Level"` to `0.5` and save: a slow swell.
7. Close Notepad: the controller rests.

**Mods** now lists **Notepad**; **Open** shows the mod's page: a level for each feel, the lightbar colours and the trigger effects.

### Change a feel

Make the tap higher. In `mods\my-data-mod\feels\tap.feel.json`, change `"f0_hz": 180` to `"f0_hz": 220`.

Save. Bururu loads the change within a second. Set the alarm flag in `state.json` again and feel the higher tap.

### Check it

```
.\brr mod check mods\my-data-mod
```

It answers `my-data-mod 1.0.0: 0 errors, 0 warnings`. A typo gets its file, line and column, and a hint:

```
my-data-mod/rules/state.rules.json:8:53: unknown name 'Alarn'; did you mean 'Alarm'
```

Bururu shows the same on its **Mods** page, and keeps the last good version running until you fix it.

### Test it

The template has a recording, `tests\demo.replay.jsonl`: made-up changes of `state.json` with their times. `mod replay` plays it through the mod without Notepad and without the controller, and compares what the mod does with `tests\demo.expect.txt`.

Run it now and it says the replay plays differently, because the tap sounds different. That change is the one you wanted, so write it into the expect file:

```
.\brr mod replay mods\my-data-mod mods\my-data-mod\tests\demo.replay.jsonl --update
```

From now on, the same command without `--update` tells you whether a change altered what the mod does:

```
.\brr mod replay mods\my-data-mod mods\my-data-mod\tests\demo.replay.jsonl
```

### Pack it

```
.\brr mod pack mods\my-data-mod --out .
```

This writes `my-data-mod-1.0.0.brr`. Anyone can install it with **Install from file...** on the **Mods** page. Before you share a mod, give it its own `id` and `name` in `manifest.json`.

That is a whole mod. Next, read how its files work: [The data template](quickstart-data.md).

## What a mod is

Bururu does the controller, the timing, all reading and the safety. A mod decides which game it is for, what to read, what the data means, and which feel, light and trigger effect each moment gets.

A mod cannot play louder than the player's **Strength**, keep the triggers stiff after it stops, or open a file or a network connection by itself. Its Lua has no `io` and no `os`. A mod that breaks is stopped, and the controller rests.

There are three kinds:

| Kind | What it is |
|---|---|
| `game` | owns one game: what to read, the rules, the feels, the settings and pages. It drives the controller while its game runs |
| `addon` | changes one or more game mods: their feels, rules, light rows and settings |
| `library` | names, feels and Lua that other mods build on |

## The mod package

A mod folder holds these files. Only `manifest.json` is required.

```
my-mod/
  manifest.json              what the mod is
  sensors.json               what it reads (sensors.md)
  sounds/*.print.json        prints of the game's sounds (quickstart-sound.md)
  rules/*.rules.json         event rules and layer rules (rules.md)
  lights.json                the lightbar, trigger, LED and mic tables (rules.md)
  feels/*.feel.json          the feels (feels.md)
  settings.schema.json       its settings and pages (sharing.md)
  lua/main.lua               Lua, when a table is not enough (quickstart-lua.md)
  data/*.json                tables its rules or Lua read
  layouts/*.layout.json      UDP packet layouts (layouts.md)
  screen/*.json, *.png       screen layouts, fonts and icon pictures (sensors.md)
  bindings/*.json            control preset layouts (sensors.md)
  demo/demo.replay.jsonl     what Play demo plays
  tests/*.replay.jsonl       recordings, each with a .expect.txt (testing.md)
  overrides/<mod id>/...     an add-on's changes to another mod (sharing.md)
  README.md                  what the mod does, for its page on GitHub
  LICENSE                    its licence, named on the mod's page
  CHANGELOG.md               a "## <version>" section per version, shown under Changes
  icon.png                   square, 64 to 256 pixels, at most 32 KB
```

A packed mod (`.brr`) is a zip of the same files. Only `.json`, `.jsonl`, `.lua`, `.md`, `.txt`, `.png`, `.wav` and `.glyphs` files and licence files are allowed. Every JSON file may have `//` comments and trailing commas. An unknown field is an error with a "did you mean" hint. Every field of every file: [Mod file reference](files.md).

## Where mods live

Bururu's data folder is the folder of `Bururu.exe`, or `%APPDATA%\Bururu` when it cannot write there: the one with `bururu.json` in it.

| What | Where |
|---|---|
| Mods | `mods\<id>\` as a folder, or `mods\<id>-<version>.brr` |
| The player's changes to a mod's settings | `mod-settings\<id>.json` |
| What a mod's sensors save | `mod-data\<id>\` |
| Which mods are on, their order, what the player allowed | `mods.json` |

A changed mod folder is loaded again within a second. A packed mod is not.

## Chapters

- [The data template](quickstart-data.md): how the quick start's mod reads, decides and lights up.
- [A mod in Lua](quickstart-lua.md): the same mod as a small state machine.
- [A racing game over UDP](quickstart-udp.md): telemetry packets and the layout that decodes them.
- [Sounds of a game](quickstart-sound.md): feels for the game's own sounds, found by their prints.
- [An add-on](quickstart-addon.md): change another mod without copying it.
- [Rules](rules.md): event rules, layer rules, situation tables, conditions.
- [Sensors](sensors.md): what Bururu can read for a mod.
- [Binary layouts](layouts.md): decoding UDP packets.
- [Feels](feels.md): the feel format, params, the motors, Bluetooth.
- [Feels in REAPER](reaper.md): design feels by ear.
- [Testing](testing.md): recordings, replays and `mod check`.
- [Sharing a mod](sharing.md): the manifest, settings and pages, permissions, add-ons, updates.
- [Publishing a mod](publishing.md): a repo, releases and Bururu's mod list.
- [Lua API reference](lua-api.md) and [Mod file reference](files.md): every function and field, generated from Bururu's code.

## Commands

Run these in PowerShell in Bururu's data folder. When that is `%APPDATA%\Bururu`, go there first (`cd $env:APPDATA\Bururu`) and type the full path of `brr.exe`.

| Command | What it does |
|---|---|
| `mod new <id> --template data\|lua\|udp\|sound\|addon` | writes a working mod into `mods\<id>`; `--game`, `--exe` and `--port` fill in the game |
| `mod check <folder>` | checks the mod as Bururu loads it; `--conflicts`, `--taste`, `--with <mod>`, `--json` add more |
| `mod replay <folder> <file.replay.jsonl>` | plays a recording through the mod and compares; `--update` writes the expect file |
| `mod record <id> --out <file.replay.jsonl>` | records a game mod's files and packets while you play |
| `mod measure <screenshot> <x0,y0,x1,y1>` | measures a region of a screenshot for a screen layout; `--icon` saves it as a picture, `--text` adds a number's digits to a font |
| `mod pack <folder>` | checks the mod and packs it as `<id>-<version>.brr` |
| `mod extract <id>` | writes a built-in or installed mod as a folder |
| `mod listing <file.brr>` | what Bururu's mod list would record of a pack, and what it would refuse |
| `mod install`, `mod update`, `mod rollback` | install from the mod list or a file, update and roll back, as the **Mods** page does |
| `mod schema <folder>`, `mod types <folder>` | write the editor support again |
| `feel import`, `feel render`, `feel check`, `feel reaper-kit` | feels from REAPER, a feel as a WAV file, the feel checks, the REAPER kit |
| `games` | the games Bururu finds on this PC, and the mods for them |
| `sound record`, `sound cut`, `sound print`, `sound find`, `sound watch` | record a game's own sound, cut a clip, make its print, find prints in a recording or in the running game ([Sounds of a game](quickstart-sound.md)) |
| `sound haptics` | what sound haptics make of a recording, as a WAV file with its numbers ([Sound haptics](player/feel.md#sound-haptics)) |
| `memory scan`, `memory next`, `memory watch`, `memory pointers`, `memory pattern`, `memory check` | find a value in the running game's memory, its pointer chain and a pattern for it, and check a memory sensor's values ([Finding a value with brr memory](sensors.md#finding-a-value-with-brr-memory)) |

Put `.\brr` in front of each: `.\brr mod check mods\my-mod`.

With **Developer tools** on (the window's **Settings** page), the **Mods** page has a **Developer** tab: it shows what the running mod reads and does, plays any feel, records a session and makes a mod from a template.

### Editor support

`mod new` and `mod extract` also write editor support into the mod folder. In VS Code, every mod file then completes its fields, and with the "Lua" extension the Lua API completes with its help. `mod pack` leaves these files out.

### On Linux

On Linux the commands run on the AppImage itself. Bururu's data folder there is `~/.config/bururu`, and paths take `/` in place of `\`:

```
cd ~/.config/bururu
~/Applications/Bururu-x86_64.AppImage mod check mods/my-mod
```

For the short `brr`, add this line to `~/.bashrc`, with the path of your AppImage, and open a new terminal:

```
alias brr='~/Applications/Bururu-x86_64.AppImage'
```

Then `brr mod check mods/my-mod` works as on Windows, and so does every other command here and every option, such as `brr -padtest`. Only `brr play` and `brr version` are Windows only: start the AppImage itself, and use `brr -version`.

## Words in this guide

| Word | Meaning |
|---|---|
| tick | one turn of Bururu's loop, every 25 ms: read, run the mods, send the controller its frame |
| sensor | Bururu code that reads something for a mod: a file, the screen, UDP packets, the game's sound |
| print | a fingerprint of a short clip of a game's sound: where its peaks are, never the sound |
| input | one thing a sensor delivers, such as `state:change` or `telemetry:gear_change` |
| snapshot | a sensor's latest values, read any time as `sensors.<id>` |
| rule | "when this input comes and this holds, play that" |
| layer | a continuous feel that plays every tick while its condition holds |
| feel | what the haptics play: a one-shot or a layer |
| frame | what the controller shows this tick: the lightbar, the triggers, the LEDs |
| overlay | a timed flash of the lightbar or the triggers |
| active | the game runs and the mod says the player is in control; only then do feels play |
| hook point | a named place in a game mod's Lua where an add-on may change what happens |
