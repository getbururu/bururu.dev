# Testing

You can test a mod without the game and without the controller. `mod check` reads every file as Bururu does. A recording holds what the sensors read during some play; `mod replay` plays it through the mod and compares what the mod does with an expect file. So after a change you know at once whether the mod still does what it did.

## Checking a mod

```
.\brr mod check mods\my-data-mod
```

`mod check` takes a mod folder, a `.brr`, or the id of an installed mod. It finds files that do not parse, unknown fields, manifest problems, feels and params that do not exist, rules on inputs no sensor delivers, conditions that do not compile, settings problems, Lua that fails to load, and files a mod may not hold. Each problem has its file, line and column. The last line sums up, and the command fails when there is an error.

| Option | Adds |
|---|---|
| `--conflicts` | what the mod changes in other mods, and who wins |
| `--taste` | the [taste lint](feels.md#the-taste-lint) on every feel |
| `--with <folder or .brr>` | a mod it builds on that Bururu does not have, such as the game mod an add-on changes |
| `--json` | the result as JSON, as the mod list's check reads it |

A check reads none of the PC's game folders, so it says the same on every PC.

## Recordings

A recording is a text file of JSON lines. The first line says what it is for; each other line is one thing a sensor read, at its time:

```
{"replay": 1, "mod": "my-data-mod", "mod_version": "1.0.0", "bururu": "0.9.0", "start_unix_ms": 1767268800000, "poll_ms": 25}
{"t_ms": 0, "sensor": "game", "running": true, "in_front": true}
{"t_ms": 0, "sensor": "state", "raw": "{\"Flags\":0,\"Level\":0}"}
{"t_ms": 500, "step": "The alarm goes on, then off"}
{"t_ms": 500, "sensor": "state", "raw": "{\"Flags\":1,\"Level\":0}"}
```

| Line | What it holds |
|---|---|
| `"sensor": "<id>", "raw": "..."` | what a file sensor read |
| `"sensor": "<id>", "b64": "..."` | a UDP packet, in base64 |
| `"sensor": "<id>", "state": {...}, "events": [...]` | what the screen sensor read |
| `"sensor": "<id>", "events": [{"name": "pickup", ...}]` | the sounds the sound sensor found, with their fields; never the sound |
| `"sensor": "<id>", "events": [{"left": 0.6, "right": 0.2, ...}]` | what the game sent to its controller, one change each (the game output sensor) |
| `"sensor": "pad", "state": {...}` | the controller: `r2`, `l2`, `held`, `sticks`, `gyro_raw` |
| `"sensor": "game", "running": ..., "in_front": ...` | whether the game runs and is in front |
| `"step": "<caption>"` | a caption; with `"reset": true` the mod starts fresh |

Sensors record what they read raw, so a changed sensor or layout decodes old recordings again. You can write a recording by hand, as the templates do, to test a moment that is hard to make happen in the game.

A recording never holds the game's sound: it is not yours to share, it may hold other people's voices, and it would be big. The sound sensor records which of its sounds it found, as the lines above show, and `mod record` leaves it out. To test a sound mod, `.\brr sound find <mod folder> <recording.wav> --replay tests\game.replay.jsonl` writes what the sensor finds in a recording of the game's sound you keep outside the mod. A changed rule or feel replays fine from it; a changed print needs a new one. Lines written by hand work too. See [Sounds of a game](quickstart-sound.md).

## Replaying

```
.\brr mod replay mods\my-data-mod mods\my-data-mod\tests\demo.replay.jsonl
```

The replay runs the mod tick by tick, with the mod's default settings and a DualSense on its cable, and compares what it does with the expect file beside the recording (`tests\demo.expect.txt`). It says `ok`, or that the replay plays differently, with the first lines that differ.

| Option | What it does |
|---|---|
| `--update` | writes what the mod does into the expect file |
| `--print` | prints it |
| `--rumble` | plays as a controller with motors only |
| `--expect <file>` | another expect file |

### Reading an expect file

```
@ 0.500
  read play tap side=both scale=1 pan=0 amount=0 at=+0  origin=rule:my-data-mod/alarm_on
  out light 255 40 0 @200
  shown light=alarm l2=- r2=-
  synth on on
audio b05053d3
```

| Line | Meaning |
|---|---|
| `@ 0.500` | a tick, at its time in seconds; quiet ticks are left out |
| `read play tap ...` | a command, with the phase that gave it |
| `origin=...` | where it came from: a rule, or a Lua file and line |
| `out ...` | what the controller shows, when it changes: triggers, lightbar colour and brightness, LEDs |
| `shown ...` | the colour and trigger keys the window shows |
| `synth on on` | the left and right actuators start or stop |
| `facts ...` | what the game mod says about the player, when it changes |
| `# ...` | a step of the recording |
| `audio b05053d3` | a fingerprint of everything the actuators played since the last step |

## Recording a session

```
.\brr mod record my-racer --out lap1.replay.jsonl --preview
```

`mod record` records a game mod's file and UDP sensors while you play, until Ctrl+C, or for a time with `--for 10m`. Start it before the game. The **Developer** tab of the **Mods** page records too, and it also records what the game sends to its controller (the game output sensor), which reaches only the running Bururu.

Recordings get shared, so they leave out what is not yours to share. A sensor's `redact` lists the fields to replace, by input name, `"*"` for every input:

```jsonc
"redact": { "*": ["PlayerName"], "ChatMessage": ["From", "Text"] }
```

A replaced text becomes `"redacted"`, numbers become 0. `--preview` shows what was replaced and asks before it writes the file.

## Tests in a mod

Keep recordings in the mod's `tests` folder, each with its expect file: `tests\lap1.replay.jsonl` and `tests\lap1.expect.txt`. Run each after a change, and all of them before a release. When a difference is what you meant, run it with `--update`.

A game mod's `demo/demo.replay.jsonl`, named by `"demo"` in its manifest, is what **Play demo** plays, with the controller live.

`-feeltest` lets a player compare the choices of a feel setting. For a setting with choices, a recording `demo/feel-<setting>.replay.jsonl` is played once per choice, "off" left out.
