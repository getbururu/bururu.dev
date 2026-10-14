# Feels in REAPER

On its cable, a DualSense is also a 4-channel audio device, and channels 3 and 4 drive its two haptic actuators. So you can design a feel in [REAPER](https://www.reaper.fm/) like any sound, feel it in your hands while you change it, and import it into your mod as feel files. You can also write feel files by hand, see [Feels](feels.md).

## The kit

```
.\brr feel reaper-kit --out reaper-kit
```

| File | What it is |
|---|---|
| `starter.RPP` | a REAPER project with two feels: `my_swell`, a one-shot, and `my_texture`, a layer |
| `bururu_voice.jsfx` | the Bururu Voice: waveform, sweep, envelope, level, sides, tremolo, gate, noise, grains. Bururu plays the same voice, sample for sample |
| `bururu_out.jsfx` | for the master track: Bururu's last step (strength, the soft limit, the controller's output level) |

## Set up REAPER

1. Copy both `.jsfx` files into a folder named `Bururu` inside REAPER's `Effects` folder. **Options -> Show REAPER resource path** opens the folder that holds `Effects`.
2. Plug the DualSense in with its cable. In REAPER's audio device settings, pick "Speakers (DualSense Wireless Controller)" with 4 outputs at 48 kHz. `starter.RPP` sends the master to outputs 3 and 4, the left and right actuators.
3. Open `starter.RPP` and play.

Quit Bururu, or pause its effects, while you design: both would play on the actuators.

## How a project becomes feels

- Each track is a feel, named as your mod plays it: `"play": "my_swell"`.
- Each Bururu Voice on a track is one voice of the feel.
- A track with one voice whose **Mode** is "Held (layer)" is a layer; other tracks are one-shots.
- Bypassed voices, and tracks without voices, are left out.
- A voice you rename in REAPER gets that name as its id; unnamed voices are `v1`, `v2`, ... in order.

## Import

```
.\brr feel import reaper-kit\starter.RPP --mod mods\my-data-mod
.\brr feel check mods\my-data-mod --taste
```

`feel import` writes `feels/<track>.feel.json` for each track, and says which are new, changed or the same; `--dry-run` only says so. Run it again after each change in REAPER: it replaces the voices and keeps what you added by hand (`params`, each voice's `mod`, `mix`, `rumble`, tags and description). It leaves alone a feel that was not made in REAPER, unless you add `--force`, and never deletes a feel whose track is gone.

`feel check --taste` warns about what tends to feel bad, see [The taste lint](feels.md#the-taste-lint).
