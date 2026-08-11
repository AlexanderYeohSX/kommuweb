# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed
- `ka2` Inside explode sequence re-exported from latest Blender PCB traces (Wave 2.15 / Trace_Less 0.30); cache-bust `?v=9`.
- `ka2` Inside caption rewritten so it doesn’t mirror the Ports “Equipped with…” structure.
- `ka2` Inside caption now matches Ports voice (8nm 8-core / Mali-G610 / 4GB / 128GB / Ubuntu 22.04 / 3× CAN FD); eyebrow and title use the same size and colour as Ports.
- `ka2` Inside callouts: 45 FOV tag nudged 1% right and 0.5% up.
- `ka2` Inside callouts: 45 FOV tag dropped 1%, leader shortened (16→11) so the text sits closer in.
- `ka2` Inside callouts: 45 FOV camera tag moved 10% left and 2% up.
- `ka2` Inside callouts: 45 FOV camera tag moved 15% left.
- `ka2` Inside callouts: added “Automotive Grade 45 FOV Camera” on the narrow lens; 185 label is now singular “Automotive Grade 185 FOV Camera”.
- `ka2` Inside callouts: CAN label copy is now “CAN Microcontroller”.
- `ka2` Inside callouts: CAN text nudged slightly up onto the leader; PCB callout dropped 1.5%.
- `ka2` Inside callouts: CAN Flexible Data is a single line so it clears the top cover; PCB leader tip pulled off the board (~28px clearance).
- `ka2` Inside callouts: CAN aligned to Aluminium’s text column/gap; PCB tip pulled off the board corner and matched to Rockchip’s line length, gap, and text column.
- `ka2` Inside callouts: PCB and CAN leaders +2% right.
- `ka2` Inside callouts: CAN leader 2% right (tip shifted, text held); PCB leader +1% right; Automotive text +2% right.
- `ka2` Inside callouts: CAN text 3% right with leader extended to the label; PCB leader +2% right; Automotive text/line 1% down, text 2% right, line 1% left then extended to the label.
- `ka2` Inside callouts: CAN leader +20% (`lineLen` 12→14.4), CAN copy split to “CAN Flexible” / “Data” and text shifted 5% left; PCB leader extended 3% right (`lineLen` 14→17).
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
