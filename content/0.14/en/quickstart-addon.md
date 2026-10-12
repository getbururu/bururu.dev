# An add-on

Back to the [modding guide](index.md).

An add-on changes another mod without copying it. It names only what it changes, so it keeps working when that mod updates, and it works beside other add-ons. The `addon` template, "Softer taps", changes the [data template](quickstart-data.md)'s mod three ways:

- taps play softer: a patch on one voice of the `tap` feel;
- the alarm going off no longer taps: the `alarm_off` rule is removed;
- the alarm light blinks slower: a patch on the lightbar's `alarm` row.

For a real game, put that game mod's id where this one says `my-data-mod`.

## Make it

First the mod to change, then the add-on beside it:

```
.\brr mod new my-data-mod --template data
.\brr mod new softer-taps --template addon
.\brr mod check mods\softer-taps --conflicts
```

`--conflicts` lists each thing the add-on changes, and who wins:

```
Feel    my-data-mod/tap            my-data-mod (base) -> softer-taps (updates, patch: voices) winner: softer-taps
Rule    my-data-mod/alarm_off      my-data-mod (base) -> softer-taps (updates, removed)       winner: (removed)
Row     my-data-mod/lightbar/alarm my-data-mod (base) -> softer-taps (updates, patch: blink_hz) winner: softer-taps
```

A patch on a name the mod does not have is a warning, and does nothing.

## How it changes the mod

The manifest says what the add-on is for, and which versions:

```jsonc
"kind": "addon",
"for": ["my-data-mod"],
"dependencies": { "my-data-mod": { "kind": "required", "version": ">=1.0.0 <2.0.0" } }
```

A version 2.0 of the mod may rename what the add-on patches, so the range keeps the add-on off until you update it. Its files under `overrides/my-data-mod/` change that mod, each by name.

A feel, by voice id (`overrides/my-data-mod/feels/tap.feel.patch.json`). The `tap` feel has one voice, `click`; the patch changes one field of it:

```jsonc
{ "voices": { "click": { "level": 0.25 } } }
```

A rule, by rule id (`overrides/my-data-mod/rules/softer.patch.json`); `null` removes it:

```jsonc
{ "alarm_off": null }
```

A light row, by row id (`overrides/my-data-mod/lights.patch.json`):

```jsonc
{ "format": 1, "lightbar": { "alarm": { "blink_hz": 1 } } }
```

The voice ids of an installed mod's feels are in its feel files: `.\brr mod extract <id>` writes the mod into `mods\<id>`. Delete that folder afterwards: a folder in `mods` replaces the installed mod.

## Feel it

Start Bururu: the **Mods** page lists **Softer taps** as an add-on. Change `state.json` as in the [quick start](index.md#feel-it): the taps are softer, the alarm going off is silent, and the light blinks slower. Switch the add-on off and on to compare.

## Change it

Make the light slower still. In `mods\softer-taps\overrides\my-data-mod\lights.patch.json`, change `"blink_hz": 1` to `"blink_hz": 0.5`.

```
.\brr mod check mods\softer-taps
```

Then write the change into the add-on's test and check against it:

```
.\brr mod replay mods\softer-taps mods\softer-taps\tests\softer.replay.jsonl --update
.\brr mod replay mods\softer-taps mods\softer-taps\tests\softer.replay.jsonl
```

The replay loads the add-on with the mods in the same folder, here Bururu's `mods` folder, so the mod it changes must be there too. `mod check` can also take that mod from somewhere else, with `--with <its folder or .brr>`; `mod replay` cannot.

## Pack it

```
.\brr mod pack mods\softer-taps --out .
```

More: [Add-ons](sharing.md#add-ons) for everything an add-on can change, and [Hook points](sharing.md#hook-points) to change a mod's Lua.
