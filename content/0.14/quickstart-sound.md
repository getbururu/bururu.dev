# Sounds of a game

Many moments of a game have a sound of their own: a pickup, a reload, an alarm, a menu click, a low-health warning. A game plays the same sound each time, so Bururu can find it in the game's sound and give the mod an input, and the mod plays a feel you designed. Such a mod needs no file or packet from the game, and often no Lua: a few prints of the game's sounds and a rule for each.

The `sound` template is such a mod. It knows one sound, a chime it made itself, and plays a short push when it hears it. For your game, you record the game, cut out its sounds and make their prints.

This works on Windows 11. Windows 10 has no way to hear one program's sound alone, and Bururu never listens to the whole PC.

## Make it

Give `mod new` the game's name and its program (Task Manager's **Details** tab shows it while the game runs):

```
.\brr mod new my-sounds --template sound --game "Lantern Isle" --exe LanternIsle.exe
.\brr mod check mods\my-sounds
.\brr mod replay mods\my-sounds mods\my-sounds\tests\demo.replay.jsonl
```

The replay answers `ok`: the mod does what `tests\demo.expect.txt` says.

| File | What it is |
|---|---|
| `manifest.json` | the game's name and program, a marker file of the game's folder, and the permission to listen to the game's sound |
| `sensors.json` | the sensor `sounds`: the sounds it knows, each with its print |
| `sounds/pickup.print.json` | the print of the template's chime |
| `rules/sounds.rules.json` | on `sounds:pickup`, the feel `thump`, as strong as the sound was loud |
| `feels/thump.feel.json` | a short 170 Hz push |
| `settings.schema.json` | the effect level on the mod's page |
| `tests/demo.replay.jsonl` | three made-up finds of the chime |
| `.gitignore` | keeps `clips/`, `recordings/` and WAV files out of a repo |

## Record the game

Start the game, then:

    .\brr sound record LanternIsle.exe --for 2m --out mods\my-sounds\recordings\game.wav

It records only the game's own sound: that program and the programs it started, nothing else on the PC. It stops when the game closes, or after `--for`. Turn off voice chat while you record: other players' voices are part of the game's sound. Play the moments you want to feel, a few times each, and write down roughly when each happened.

## Cut the sounds

    .\brr sound cut mods\my-sounds\recordings\game.wav --at 41.2 --out mods\my-sounds\clips\pickup.wav

`--at` finds the sound around that time, from its start until it falls 30 dB below its loudest, at most 2 seconds. `--from 41.15 --to 41.60` cuts by hand. Listen to the clip: it should start where the sound starts, as the start is what Bururu hears first. The cleaner the clip (no music or other sounds under it), the better its print.

## Make the prints

    .\brr sound print mods\my-sounds\clips\pickup.wav --out mods\my-sounds\sounds\pickup.print.json

It says how many peaks the print has and how findable the sound is. A sound with a clear pitch or several parts works best. A noise burst, a very short tick or a steady hum have few peaks that stay the same; pick another moment of the game for them, or a longer clip.

A sound the game plays in a few versions (three reload sounds that take turns) gets a print for each, under one name: `"reload": { "prints": ["sounds/reload-1.print.json", "sounds/reload-2.print.json"] }`. A sound the game plays higher or lower each time gets `"pitch": 1` or `2`.

## Name them and give them feels

In `sensors.json`, one entry per sound:

```jsonc
"sounds": {
  "pickup": { "print": "sounds/pickup.print.json" },
  "alarm":  { "print": "sounds/alarm.print.json", "min_match": 0.7 }
}
```

In `rules/sounds.rules.json`, one rule per sound, with a feel from `feels/` ([Feels](feels.md)):

```jsonc
"alarm": { "on": "sounds:alarm", "play": "alarm_pulse" }
```

`mod check` warns about a sound no rule listens to.

## Try it on the recording

    .\brr sound find mods\my-sounds mods\my-sounds\recordings\game.wav

It lists each place a sound is found, with its match, strength and delay. Each moment you played should be there, and nothing else. Record a few minutes of the game's music and menus without those moments too, and run `sound find` on that: a sound made of notes can be taken for music that plays the same notes. When a sound is found where it is not, raise its `min_match`; when it is missed, lower it, or cut a cleaner clip.

To test a change of the rules or the feels without the game:

    .\brr sound find mods\my-sounds mods\my-sounds\recordings\game.wav --replay mods\my-sounds\tests\game.replay.jsonl
    .\brr mod replay mods\my-sounds mods\my-sounds\tests\game.replay.jsonl --update

The replay holds what was found, never the sound. See [Testing](testing.md).

## Feel it

    .\brr sound watch mods\my-sounds

Start the game: it prints each sound it finds, live, and plays nothing. Then turn the mod on in Bururu and play.

While the mod list does not name the game's program, Bururu listens to the game only with **Developer tools** on (the window's **Settings** page), and only where the manifest's `markers` are in the game's folder or above it. The template's marker is the game's program, which does not count, as every program of that name has it: put a file or folder of the game's own in `markers` first, such as its data folder (`"markers": ["LanternIsle_Data"]`). The log says when the marker is missing. For a published mod, the mod list's record of the game names its program.

## What stays out of a mod

- **The game's sounds.** A mod ships prints, never the clips or the recordings. They are the game maker's, and a recording may hold other people's voices. Bururu leaves `clips/` and `recordings/` out of the mod and out of `mod pack`; a WAV file in `sounds/` is an error, elsewhere a warning.
- **Other programs.** The sensor listens to the game's own programs only. Bururu also keeps a list of programs it never listens to, whatever a mod says: browsers, voice and chat apps, music and video players, streaming, recording and remote-play tools, and launchers.

## What it does not do

- The input comes about 100 ms after the sound starts, sometimes more: Bururu must hear its start first. The feel then still has the way through the haptics stream to go. On the cable the haptics play through a shorter buffer while the sensor listens, and the log says once how long a push takes there (`Native haptics: through WASAPI a push takes about ...`). That suits pickups, reloads, alarms and menus. For the first moment of a shot or an explosion, it is late.
- A game that keeps the sound device to itself (exclusive mode) shares no sound, and the log says so after a minute of silence.
