# Effects

What you feel comes from the game mod: which moment plays which feel, light and trigger effect. Bururu plays it on the controller and keeps your limits. This page covers what is the same for every game.

## Haptics

The DualSense has two haptic actuators, one in each grip. Bururu plays a mod's feels on them as sound: on the cable through the controller's own audio device, on Bluetooth inside the reports it sends. The feels are made with Bururu's synth, so a mod can play clicks, swells, rattles and textures, each on the left, the right or both.

- **Strength**, on **Home** under **Outputs**, scales every effect of every game, from 0% to 300%. No mod plays louder.
- **Haptics mode**, under **More controller settings** on **Home**: **Auto** plays the feels on the actuators. **Rumble** plays them on the two rumble motors instead, as older controllers do. Rumble mutes the actuators, so Bururu rumbles only with **Rumble**.
- **Effect levels**, on a game's page, set each of its mod's effects: 100% is as the mod made it, 0% turns it off.

## Sound haptics

In a game without a mod, Bururu can listen to the game's sound and turn it into haptics, in one of two modes:

- **Follow** (the default): the haptics follow the game's sound as it plays, music and voices too. You feel the game's own sound from about 115 to 300 Hz as it is, a band the actuators play cleanly. The deep bass under about 100 Hz, where the actuators rattle, plays as a smooth 160 Hz tone that follows its level (**Bass**), so a steady bass note or engine hum never reaches the actuators as it is. In between, a sound plays partly as each. A sound that rises well over the sound before it gets a lift, and steady sound such as music, an engine or rain is turned down, more so after a few seconds of it, so hits stand out (**Punch**). Sound far under the game's usual level fades out (**Sensitivity**), and silence is silent.
- **Impacts**: a short push when something hits: a shot, an explosion, a thump. Each push is a soft swell between 140 and 190 Hz that Bururu plays itself; you never feel the game's sound as it is. Music that keeps a beat, speech, and steady sound such as an engine or rain give no push.

Both follow the game's own level, so they feel the same at any game volume. In Follow the game's usual level is taken from the last 3 seconds of its sound, so a single hit or a short pause does not change it and a footstep stays smaller than an explosion. After a lasting change, such as from a menu into a fight, it settles within 2 to 4 seconds.

- **Off at first.** Turn them on for one game on its page, under **Sound haptics**, or for every game without a mod on the **Controller** page under **Sound haptics**, where the mode and their other settings are. Both say whether Bururu listens now, and why not when it does not. To try them for one run: quit the Bururu in the tray, then run `.\brr -console -soundhaptics -verbose`. With `-verbose` it prints each push in Impacts, and in Follow the signal's peak and level once a second.
- **Which games**: the games the mod list has a profile for, while no game mod runs here, and a program you name for one run with `-soundgame`. A game's entry in `games` turns them on or off for that game.
- **When**: while Bururu has the controller. With the effects paused, with a game that reads the DualSense itself (Direct), or with a game that plays haptics of its own, they stay off.
- **Strength** scales them like every other effect, and the same limit holds. In Follow, a Strength over 100% makes the quieter parts stronger, while the strongest parts stay where they are at 100%, since more would distort and Bluetooth would compress them. On a controller without HD haptics, or with **Mode: Rumble**, they play on the motors: in Follow the motors follow the signal's strength, the big one its deep bass, and stay still while it is faint.
- **Windows 11 only**: Windows 10 cannot share one game's sound. On Linux and macOS they stay off too.
- In a game with surround sound, Follow takes the main channels and the low-effects (LFE, subwoofer) channel together. Impacts listens to the LFE channel alone, where explosions and hits are and music and speech mostly are not; the log says so.
- **On the cable** the haptics play through a shorter buffer while sound haptics run, so they land sooner. How much sooner depends on the controller's sound device; the log says once how long a push took on each (`Native haptics: through WASAPI a push takes about ... ms to the actuators, through waveOut about ... ms`). They move to it and back with a fade of a few milliseconds, and Bururu's own effects lose nothing at a move. Where Windows does not start the shorter buffer, the haptics keep the usual one and the log says why. In Follow the signal waits in a ring of its own on its way to the haptics, as the game's sound does for the headphones: just long enough that it never runs dry, about 22 ms with Windows' usual 10 ms periods. On Bluetooth add the 10 to 40 ms Windows' Bluetooth takes.
- Once, 10 seconds after it starts listening, the log says how long it takes from the game's sound to the controller.
- **To see what they make of a sound**: record the game with `brr sound record`, then `brr sound haptics game.wav --out haptics.wav` plays it through sound haptics as on the cable, writes what the actuators get, and prints its peak, its level, how much of it is under 100 Hz and how much is silent. Its options are the settings (`--mode`, `--level`, `--bass`, `--punch`, `--sensitivity`, `--side`, `--strength`); `--bluetooth` writes it as Bluetooth carries it.

**What it listens to**: only the sound of the game's own programs (the game, and the programs it started), while the game runs, through Windows' process loopback. Never the whole PC and never your microphone. Bururu also keeps a list of programs it never listens to, whatever a setting says: browsers, voice and chat apps, music and video players, streaming, recording and remote-play tools, launchers and Windows itself. Bururu looks at the sound as it comes and keeps nothing; nothing is sent anywhere. On a Windows without process loopback it listens to nothing.

## Trigger effects

A game mod gives R2 and L2 an effect for each situation, such as a click while you aim. You change them on the mod's **Triggers** page. Each card has a feel, the trigger's travel from 0 (at rest) to 9 (pressed all the way) with handles where the feel starts and ends, and a strength from 1 to 8.

| Feel | What you feel | Mode in the settings file |
|---|---|---|
| **Free** | No resistance | `OFF` |
| **Click** | Resistance that gives way with a click, like a gun trigger | `WEAPON` |
| **Wall** | Resistance from a point to the end | `FEEDBACK` |
| **Rattle** | A vibration from a point to the end, with a speed from 1 to 40 | `VIBRATION` |
| **Ramp** | Resistance that grows or falls from one point to another | `SLOPE_FEEDBACK` |
| **Wall by position** | A resistance for each of the 10 positions | `MULTIPLE_POSITION_FEEDBACK` |
| **Rattle by position** | A vibration for each of the 10 positions | `MULTIPLE_POSITION_VIBRATION` |

On an Xbox One or Xbox Series controller on Windows the effects play on its trigger motors instead, as short pushes and soft ticks. How hard they push, and whether they tick, you set under **More** in that controller's settings on the **Controller** page.

**Exact values** on a card shows every value of the effect, the by-position feels too, and **Back to default**.

**Try on R2** (or L2, or both) plays the card's effect for 2 seconds, then frees the trigger. It works while Bururu has the controller and no game drives it: with no game running, in a game's menu, or with the effects paused. It needs **Adaptive triggers** on. When it cannot play, a line under the button says why.

## Lights

A game mod sets the lightbar's colour for each state of the game, the five player LEDs and the mic LED. You pick the colours on the mod's **Lightbar colours** page. **Brightness**, and switches for the lightbar, the player LEDs and the mic LED, are on **Home** under **Outputs**.

## Gyro aim

With gyro aim on, Bururu turns the controller's rotation into mouse movement, on the desktop and in any game that aims with the mouse.

- **When it aims**: **Always**, **While you hold** a button (L2 at first), or **Press** a button to turn it on and off (the touchpad at first). It stays still while a finger rests on the touchpad, and while a game mod says a menu is open (**Pause in game menus**).
- Turning the controller sideways moves the mouse sideways, and rolling it adds to that (**Roll**, 60% at first). Tilting it moves the mouse up and down. **Sideways** and **Up and down** scale each: at 1, about 22.5 mouse counts per degree.
- **Slow movement: Steady** (the default): very slow movement moves nothing, so the aim holds still. **Exact** lets every bit through: finer aim, but the gyro must be well calibrated. A DualSense Edge suits **Exact** better.
- Gyros drift a little. Bururu learns the drift only while the controller lies still for 2 seconds, never while you hold it, and keeps it when the controller connects again, also when it moves from Bluetooth to its cable or back. It keeps the DualSense's in `gyro_calibration_direct.json`, and a DualShock 4's until Bururu quits.
- If the game runs as administrator, Windows keeps Bururu's mouse movement from it: start the game normally, or run Bururu as administrator too.
- `.\brr -gyrotest` measures the gyro on the cable and saves a calibration.

## When Bururu lets go

In a game's menu and with no game running, the player LEDs show the controller's number, and while **Adaptive triggers** and **Lightbar** are on, the triggers are free and the lightbar glows dim in its colour (blue, or the one picked on **Controller**). When you pause and when Bururu quits, the controller goes back to rest: the LEDs off, the haptics silent, and while **Adaptive triggers** and **Lightbar** are on, the triggers free and the lightbar blue. On Linux the player LEDs, and a DualShock 4's lightbar, go back to what Linux showed for that controller. Bururu writes only the parts it controls: with **Adaptive triggers** or **Lightbar** off on **Home**, the triggers or the lightbar stay as the game or another program set them.

If Bururu crashed, the triggers may stay stiff: start Bururu again, or turn the controller off and on.
