# A mod in Lua

The `lua` template is the [data template](quickstart-data.md)'s mod with its logic in Lua: a small state machine that knows "calm", "alarm", and two seconds of "cooling" after an alarm. Use Lua when a moment depends on what came before: a timer, a count, a value that rises and falls. A mod can mix rules and Lua, see [When to use Lua](rules.md#when-to-use-lua).

## Make it

```
.\brr mod new my-lua-mod --template lua
.\brr mod check mods\my-lua-mod
```

The folder holds the data template's `manifest.json`, `sensors.json`, feels and settings, with `lua/main.lua` in place of the rules and `lights.json`. The manifest names it: `"entry": "lua/main.lua"`. `mod check` also loads the Lua, so a syntax error shows here with its line.

## How the Lua works

Bururu runs `lua/main.lua` once when the mod loads. The file keeps what it knows in `bururu.state`, and hands Bururu functions to call at their place in each tick:

```lua
local s = bururu.state
s.mode = "calm"     -- "calm", "alarm", or "cooling" for two seconds after it
s.cool_until = 0    -- when cooling ends, in ns since the session started
```

Times are nanoseconds since the session started, so 2 seconds is `2e9`.

`bururu.on` runs a function for each input it names, here every change of `state.json`:

```lua
bururu.on("state:change", function(ev)
  local alarm = ev.new.flags.Alarm
  if alarm and s.mode ~= "alarm" then
    set_mode("alarm")
    feel.play("tap")
  elseif not alarm and s.mode == "alarm" then
    set_mode("cooling")
    s.cool_until = ev.at + 2e9
    feel.play("tap", { side = "left" })
  end
  ...
end)
```

`bururu.facts` tells Bururu whether the player is in control, and what to show as the game's state on the mod's page:

```lua
bururu.facts(function(t)
  return { active = true, menu = false, context = s.mode }
end)
```

`bururu.on_tick` runs every tick while the mod drives. A layer must be set every tick, or it fades out:

```lua
bururu.on_tick(function(t)
  if s.mode == "cooling" and t.now >= s.cool_until then
    set_mode("calm")
  end
  if s.level > 0 then
    feel.layer("swell", "swell", gomath.min(1, s.level))
  end
end)
```

`bururu.on_frame` sets the lightbar and the triggers, by the keys of the mod's colour and trigger settings:

```lua
bururu.on_frame(function(t, f)
  if t.now < s.flash_until then
    f:light("flash", 1)
  elseif s.mode == "alarm" then
    f:light("alarm", kit.blink(t.unix_ms, 2) and 1 or 0)
  ...
  if s.door then
    f:trigger("l2", "soft")
    f:trigger("r2", "firm")
  end
end)
```

`status.fact("mode", mode)` shows the mode on the mod's page. Every function: [Lua API reference](lua-api.md). Where each one runs in a tick: [The phases of a tick](rules.md#the-phases-of-a-tick).

### More than one controller

With several controllers connected, the feels and lights play on the one in your hands, the one used last. `t.slots` lists the connected controllers in the order of the **Controller** page, each with its `name` and `caps`, and `t.slot` is the number of the one in your hands. A mod can play on another one by its number:

```lua
for n = 1, #t.slots do
  log.info("Controller %d: %s", n, t.slots[n].name)
end
feel.play("tap", { slot = 2 })   -- on Controller 2 only
bururu.on_frame(function(t, f)
  f:slot(2):light("alarm")       -- Controller 2's lightbar
end)
```

A controller that gets no frame from the mod rests. The tables and overlays of `lights.json` set the frame of the controller in your hands.

When two people play side by side, one controller each, a game mod says how many people play in its facts, and whose feels a play is:

```lua
bururu.facts(function(t)
  return { active = true, context = s.mode, players = s.players }
end)
feel.play("hit", { player = ev.player })
bururu.on_frame(function(t, f)
  f:player(2):light("hurt")
end)
```

While the facts say two or more people play and two or more controllers are connected, player n's feels and frame go to Controller n; a player with no controller connected gets none. Controller 1, player 1's, is then the controller in your hands, and pressing buttons does not move it: the tables of `lights.json`, `f` and the feels with no player go to it. Otherwise every player's feels play on the controller in your hands.

### What Lua cannot do

The Lua is Lua 5.1 with `string`, `table` and `math` and Bururu's tables. There is no `io`, `os`, `debug`, `load` or `dofile`: a mod reads only through its sensors. `require("name")` loads `lua/name.lua` from the mod's folder.

Each call has a budget of time and memory. A call past it is stopped and its commands are dropped. A function that keeps failing is switched off for the session, with a note on the **Mods** page.

## Feel it

Open `mod-data\my-lua-mod\state.json` in Notepad and change it as in the [quick start](index.md#feel-it). After an alarm the lightbar stays a brighter blue for two seconds, and the mod's page shows **Mode: cooling**.

## Change it

Make the cooling last four seconds. In `mods\my-lua-mod\lua\main.lua`, change `s.cool_until = ev.at + 2e9` to `s.cool_until = ev.at + 4e9`.

Bururu loads the change within a second, and the mod starts fresh. To keep its state across a reload, hand it over in `bururu.on_reload(function(old_state) ... end)`.

```
.\brr mod check mods\my-lua-mod
```

An error while the mod runs shows in the log and on the **Mods** page with its file and line:

```
[my-lua-mod] Lua error in on(state:change)@my-lua-mod/lua/main.lua:20: my-lua-mod/lua/main.lua:24: attempt to call a non-function object
```

With VS Code and its "Lua" extension, every Bururu function completes with its help. After a Bururu update, `mod types` writes the help again:

```
.\brr mod types mods\my-lua-mod
```

## Test it

The replay now says the mode turns calm later than the expect file says. That is your change, so write it:

```
.\brr mod replay mods\my-lua-mod mods\my-lua-mod\tests\demo.replay.jsonl --update
.\brr mod replay mods\my-lua-mod mods\my-lua-mod\tests\demo.replay.jsonl
```

Lines from Lua name the file and line that gave the command, such as `origin=lua:my-lua-mod/lua/main.lua:24`. More: [Testing](testing.md).

## Pack it

```
.\brr mod pack mods\my-lua-mod --out .
```

Let other mods change your Lua without copying it: [Hook points](sharing.md#hook-points).
