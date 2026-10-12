# Publishing a mod

Players get your mod from **Browse** on the **Mods** page once it is in Bururu's mod list, or with **Install from file...** from a `.brr` you publish. This chapter covers both: a repo of the mod's own, releases, and the list.

The list is [getbururu/mod-index](https://github.com/getbururu/mod-index). It holds one small file per mod; each mod stays in its author's repo. The list checks every new release, a maintainer reviews it, and Bururu reads the published list. Bururu installs nothing by itself: the player clicks **Install**, and updates wait for the player's click.

## Your mod's repo

One public GitHub repo per mod, named `bururu-<id>` after the mod's id, such as `bururu-gravel-rally`. Give it the topic `bururu-mod` (the gear next to **About** on the repo's page), so people find it on GitHub. The mod's files sit at its top, as `mod new` writes them:

```
bururu-my-mod/
  manifest.json, sensors.json, rules/, feels/, ...   the mod
  README.md                       what the mod does, for its page on GitHub
  CHANGELOG.md                    a "## <version>" section per version, newest first
  LICENSE                         the licence the manifest names
  icon.png                        the mod's icon
```

- Keep your local copy in a folder named after the mod's id, as Bururu loads it: `git clone https://github.com/<you>/bururu-my-mod mods\my-mod`. Bururu skips `.git` and `.github`, so the clone loads as it is.
- Everything else at the top of the repo is part of the mod, so it must be an allowed file. A build script or an `.editorconfig` there makes `mod check` refuse the mod.
- Never commit `.brr` files: they go into releases.

`README.md` is for your repo's page on GitHub, which players open from the mod's page in Bururu; Bururu itself does not show it. Say what the mod does, which game it is for, every folder it reads and why, and why it reads the screen, the keys or a UDP port when it does. A reviewer reads it against the mod.

`CHANGELOG.md` shows on the **Changes** tab of the mod's page, from its first version heading on: a title or a note before it is for you, not for players. It shows its headings, lists, paragraphs, tables, quotes and code. A link shows its words, and images and raw HTML show nothing. A mod installed from a file shows the `CHANGELOG.md` in its own package.

The page names the licence from `license`. A standard one, such as `MIT` or `Apache-2.0`, shows only by its name; any other licence file is there to read on the page.

## Names

The list checks these in the manifest. `mod listing`, below, shows each one before you publish.

- `id`: first come, first served, and it never changes once listed. Some ids are kept: `bururu`, `brr`, `core`, `official`, `index`, `mod-index`, the list's own mods and the templates.
- `name` and each of `authors`: plain ASCII, 1 to 64 characters. Names that look like another mod's or author's are refused, look-alike letters too (0 and o, 1 and l, rn and m). "Bururu", "getbururu" and "official" are kept for the list's own mods.
- `summary`: one line of at most 200 bytes.
- `version`: `major.minor.patch`, or with a pre-release part for a test version (`1.3.0-beta.1`). Players see test versions only with **Show test versions** on.
- `license`, with the licence file beside the manifest.
- The game's id: use the list's id for your game when it has one (`games/<id>.json` in mod-index); for a new game, propose a record with your entry. The templates call their game `game`: rename it before you list a game mod.

The icon: `icon.png`, a square PNG of 64 to 256 pixels, at most 32 KB. Bururu keeps only its pixels. Without one, the manifest's `icon` names one of the window's own (`ship`, `car`, `waves`, `gear`, `mods`, `file`).

## Releasing a version

1. Set `version` in `manifest.json`, and add its section to `CHANGELOG.md`:

   ```
   ## 1.1.0

   Softer kerbs. The gear clunk follows the clutch.
   ```

   The section is the release notes, and what Bururu's update dialog shows (at most 2 KB).

2. Check, test and pack it, in Bururu's data folder:

   ```
   .\brr mod check mods\my-mod
   .\brr mod replay mods\my-mod mods\my-mod\tests\lap.replay.jsonl
   .\brr mod pack mods\my-mod --out dist
   .\brr mod listing dist\my-mod-1.1.0.brr
   ```

   `mod listing` shows what the list will record of the pack (names, permissions, games, files, icon and notes) and names anything it refuses.

3. Commit, tag the commit `v<version>`, and push the tag:

   ```
   git tag v1.1.0
   git push origin v1.1.0
   ```

4. Make a GitHub release for the tag with one file, `<id>-<version>.brr`, as `mod pack` wrote it, and the version's changelog section as its notes. On the repo's **Releases** page: **Draft a new release**, pick the tag, attach the file, paste the section. Or with the GitHub CLI:

   ```
   gh release create v1.1.0 dist\my-mod-1.1.0.brr --notes-file notes.md
   ```

Once a version is listed, never change its `.brr`; publish a new version instead.

## Getting listed

1. Publish a release as above.
2. Fork [mod-index](https://github.com/getbururu/mod-index), add `entries/<id>.json`, and open a pull request:

   ```json
   { "entry": 1, "id": "gravel-rally", "repo": "your-login/bururu-gravel-rally", "official": false, "prerelease": false }
   ```

   `repo` is your repo on github.com, and `prerelease: true` lists test versions too. The games, authors and permissions come from your manifest.

3. The pull request's check finds your repo and its newest release, and says what it found.
4. A maintainer merges. Within six hours the list runs Bururu's checks on your release and opens a pull request with a report. A maintainer reviews it and merges, and your mod is in **Browse**.

The full guide, with the reviewer's checklist: [Listing a mod](https://github.com/getbururu/mod-index/blob/main/docs/listing.md).

## Updates

Publish a new release; there is nothing to send. The list finds it within six hours and opens a pull request with what changed. A version that reaches more (a new folder, the screen, the keyboard, a new port, a new required dependency) gets the full review again, and Bururu asks the player before it updates.

- A player's download must match the size and SHA-256 the list recorded, and may ask for no more than the list said.
- Players can roll back to the two versions before, with their settings. When you rename a setting or a feel, ship a [migration](sharing.md#updates-and-migrations).
- A mod whose repo changes owner, or whose listed `.brr` changes or goes away, is held until a maintainer has looked. A mod whose repo is archived or gone leaves **Browse**; players keep it.
- A version found to do harm is revoked: Bururu warns its players, or switches it off at once for malware.

## Dependencies and add-ons

- A required dependency must be listed too, or built in. The install dialog shows it and installs it with the mod.
- **Browse** shows an add-on to players who have the game mod it changes. Libraries come as dependencies.
- An add-on is checked against the mod it changes: `.\brr mod check mods\my-addon --with <that mod's folder or .brr>`.

## What players see

- **The card**: the icon, the name, the authors and your GitHub login, the summary, the games, downloads (GitHub's count for your `.brr` files), stars of your repo, the date of the newest version, and what the mod reads, in red when it sends data off the PC. **Official** marks the list's own mods only.
- **The mod's page**: the summary, what the mod adds, every folder and port it reads, its details with a link to your repo, your changelog and its versions, with **Star on GitHub** and **Report this mod**.
