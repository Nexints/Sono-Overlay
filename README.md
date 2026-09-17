# Sono-Overlay / Sonolus Music Video creation tool

Sonolus Overlay tool to add Music Video support and other extended gimmicks / skins / etc.

> [!NOTE]
> I am not affiliated with the original developer(s) in any way, shape or form.
>
> All assets are made by myself.

This is a custom fork of [TootieJin](https://tootiejin.com)'s fork of [Nanashi's Overlay Software](https://github.com/sevenc-nanashi) (taken down) with non-Project Sekai assets, intended to provide my custom Rhythm Game skin towards Sonolus custom charts.

Sono-Overlay currently supports the following servers:
- ProSeka Servers:
  - Sekai Rush (sekai-best-)
  - SSS (sss-)
- Custom ProSeka Servers:
  - Chart Cyanvas (chcy-)
    - 0: Archive
    - 1: Offshoot
  - Potato Leaves (ptlv-)
  - Untitled Sekai (utsk-)
  - Untitled Charts (unch-)
  - Next Sekai (coconut-next-sekai-)
  - ScoreSync (sync-)
- Official Game Servers:
  - Sonolus Horizon (coconut-horizon-) / NEW Expansion (BETA)
- Local File Support (*.json.gz)
- More coming!

Sono-Overlay has these extended features compared to the originals:
- Dynamic Stage Support
- Sonolus Horizon Support
- Custom Watermarks (& Moving Watermarks)
- Primary [AviUtl ExEdit 2](https://mints.ne.jp) Support
- Local File Support (*.json.gz)

Sono-Overlay is integrated into this repository!
- [Sono-Utils](https://github.com/Sono-Suite/Sono-Utils/tree/main)
It's recommended to use SonoUtils instead of Sono-Overlay, as Sono-Utils holds Sono-Overlay bundled inside of it. However, Sono-Overlay runs standalone, which makes Sono-Utils compliant under the AGPL-v3 (as these are seperate apps, and Sono-Utils only launches Sono-Overlay)

Sono-Overlay has these features planned:
- High:
  - Custom ID (for a custom server)
  - Bundled Local Server
- Low:
  - Skill Cards
  - More Addons

Sono-Overlay will not implement these features:
- Score Maker Support
- AviUtl support (AviUtl ExEdit 2 is better)
- v1 UI support (legacy support for a UI that no longer exists)

## How to use

Required programs:
- [AviUtl ExEdit 2](https://mints.ne.jp)
- LSMASH WORKS
- MP4 Exporter
- Powershell

Required Clips:
- Sonolus Gameplay with white background
- Sonolus Gameplay with black background
- Remember to hide the UI!

Additional Notes:
- English language for AviUtl ExEdit 2 is selectable in `設定 > 言語の設定 > English` ( `Settings > Language > English`)
- If you're using AviUtl ExEdit 2, import the `AviUtl2\lwinput.aui2` and `MP4Exporter.auo2` plugin to `C:\ProgramData\aviutl2\Plugin` folder
- The recommended font is RodinNTLG DB + EB.
- System locale must be `Japanese (Japan)`, and `Japanese Language Pack` must be installed. This is a restriction with the AviUtl ExEdit program.
- While the program will work without the Japanese Language Pack or the Japanese System Locale, **I hold no guarantees the output of this program will be usable.**
- For the clips: Please make sure all clips are passed through FFMPEG or your video editor of choice. AviUtl ExEdit 2 has a syncing problem and visual bugs if you record on IOS.

### Step-step guide:
0. Install required programs
1. Download the latest version of Sono-Overlay
2. Unzip all files in a directory without non-ACII characters, and make sure this directory does not require admin permissions.
4. Once the programs are installed, run the `Sono-Overlay.exe` file, and follow the steps given by the program.
**Note: Select AviUtl ExEdit2 (`2`).** AviUtl (normal) is kept for legacy reasons.
5. Create a new project in AviUtl ExEdit 2
6. Import the .object or the .exo file from the `{Sono-Overlay Directory}\dist\[Chart ID]` into the project by drag and dropping.

This video element uses `Unmult`. If your video element does not have this, you are looking at the wrong `Video file`.

7. Click the bottom `Video File` (just above the `root@sono-overlay`), click `File`, and then select your `Black BG Video`. 

This video element uses the blending mode `Multiply`. If your video element does not have this, you are looking at the wrong `Video file`.

8. Click the `Video File` just above the Step 7 video file, click `File`, and then select your `White BG Video`. 

Step 8 is needed to ensure Dynamic Stage support.

9. Make sure both videos in steps 7 and 8 are synced, and adjust the video / audio positioning as necessary.
10. If needed, go to the `Root@sono-overlay` file, and adjust the `Offset`.
11. When finished, export your video as mp4. (Assuming AviUtl ExEdit2: `File > Export > MP4 Exporter (by えすご/Esugo)`)

Play around with this tool, and tweak the settings as you wish~

## Terms of Use

1. **(REQUIRED)** You must make this text visible in the video description or video itself.

**EN**
```
Sono-Overlay:
- Made by Nexint (https://nexint.ca/)
- Forked from TootieJin (https://tootiejin.com) and Nanashi (https://sevenc7c.com/)
- Assets / Skins by Nexint (https://nexint.ca/)
- https://github.com/Sono-Suite/Sono-Overlay
```

2. This tool **should not be used for malicious purposes** (such as, but not limited to: spreading misinformation on social media).
3. The author **assumes no responsibility whatsoever** for any issues or disadvantages arising from the use of this tool. This tool is provided AS IS, with no warranties.
4. (NEW) You are not allowed to use this tool with official SEGA assets.
5. (NEW) You are not to use this tool on other people's charts, unless you have express permission from said person.

## Legal:
The included ProSeka Faithful UI is governed under my [Nexint TOS](https://nexint.ca/tos).
The program itself is under the AGPL-v3, as this is a fork of a program which is under the AGPL-V3.

As of writing, this program is a hard fork of v0.5.5-b of pjsekai-overlay-APPEND. This is subject to change.

Please check the [Nexint TOS](https://nexint.ca/tos) when using this tool with ProSeka Faithful!

### AI Disclosure:
Some parts of Sono-Overlay were coded with AI.

Please forgive my lack of knowledge on Golang, AviUtl ExEdit2 syntax, etc!

I've made sure to independantly tweak and test everything this program outputs.