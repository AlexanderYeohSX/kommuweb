# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed
- `ka2` Inside scrub lag strengthened (desktop lerp 0.18→0.07, mobile 0.22→0.09) for GSAP-style catch-up.
- `ka2` Inside scrub: explode finishes by ~65% of pin then holds open; pin height 280→360vh; labels appear at 0.58.
- `ka2` Inside explode scrub sequence upgraded to **60 frames** (`d_01–60` / `m_01–60`) with matching posters and `data-frames="60"`.
- `ka2` Inside media keeps each frame file’s native aspect ratio and fits responsively in the stage (no stretch/crop).
- `ka2` Inside callouts: Panoramic Cameras label nudged slightly higher.
- `ka2` Inside callouts: Aluminium shorter/closer to body; cameras lowered.
- `ka2` Inside callouts: Aluminium line longer, Rockchip further out, cameras moved up.
- `ka2` Inside callouts: PCB up-right, Aluminium longer line, Rockchip lower/clear of casing, cameras lower.
- `ka2` Inside callout tips pulled ~18px short of parts and retargeted (PCB left edge, unibody front-left rim, Rockchip right edge, right camera).
- `ka2` Inside callout tips pinned to measured component edges; added ~12px gap between leader line and label text.
- `ka2` Inside callouts remapped in media coordinates from measured d_26.jpg tip positions (equal 12% leaders); removed stage-gutter auto-placement that caused misalignment.
- `ka2` Inside explode section restyled to match Canva mock: open black stage (no card frame), lighter callout type, equal-length leader lines with tip dots, L/R label layout (PCB / Aluminium Unibody / Rockchip RK3588 6 TOPS / Panoramic Cameras), and bottom scroll cue.
- `ka2` Inside callouts moved onto stage gutters outside the media box (auto-placed) so label text never overlaps the model; leaders use equal pixel length with tip clearance.
