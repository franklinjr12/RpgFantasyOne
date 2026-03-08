
# Step 12 Refinement - UI/UX (Must Be Playable and Clear)

## Scope for this refinement
- Implement backlog step `12) UI/UX (Must Be Playable and Clear)` with incremental, testable tasks.
- Prioritize skill and item icon integration using `resources/sprites/raven_fantasy_icons_32x32.png`.
- Keep current gameplay behavior unchanged; this is a presentation/UX pass.
- Keep Windows + raylib-go assumptions.

## Current baseline (from code audit)
- Main menu, class select, reward selection, and results screens exist but are mostly text-only.
- In-run HUD has text readouts and a 4-slot skill bar with placeholder gray icon boxes.
- Skill cooldown overlays and key labels already exist in `systems.DrawSkillBar`.
- No icon atlas integration exists yet for skills/items/effects.
- No minimap, buff/debuff tray, target frame, floating combat text, or pause menu.

## Definition of done for this step
- Skill icons render in HUD from the new spritesheet atlas.
- Reward UI shows item icons (at least slot-based placeholders from atlas).
- Class select and/or skill preview UI shows class skills with icons.
- Buff/debuff tray renders active effects with icon + remaining duration.
- Target frame appears when hovering/locked target, with HP and active debuff icons.
- Menus/HUD remain readable at 1600x900 and still playable with current controls.

## Task backlog

### A) Icon atlas foundation (spritesheet import done correctly)
- [x] Add a dedicated icon spritesheet asset key and load `resources/sprites/raven_fantasy_icons_32x32.png` during boot.
- [x] Introduce icon-atlas constants/utilities (tile size `32x32`, sheet size `512x4384`, columns `16`) and a helper that converts `(col,row)` to `rl.Rectangle`.
- [x] Add a safe icon draw helper that falls back to a colored rect if texture failed to load.
- [x] Keep icon atlas utilities separate from humanoid character spritesheet helpers to avoid coupling.
- [x] Add tests for atlas rect calculation (origin tile, middle tile, edge tile) to prevent off-by-one errors.

### B) Skill icon mapping and HUD integration
- [x] Create explicit skill-icon mapping for all 12 `gamedata.SkillType` values using atlas cell coordinates.
- [x] Update `systems.DrawSkillBar` to render mapped skill icons in each slot instead of gray placeholder fills.
- [x] Keep existing key label and cooldown overlay behavior; ensure overlays remain visible over icons.
- [x] Add mana/cost indicator per skill slot (small numeric or pip marker) with clear coloring when insufficient resource.
- [x] Add tests that validate every class skill resolves to a non-empty icon mapping.

### C) Item icon mapping for reward UI
- [x] Add item icon resolution (minimum: by `ItemSlot`, optional: overrides by item name for better flavor).
- [x] Update reward selection screen to draw item icons next to item title/description.
- [x] Keep reward selection readability and current keyboard flow (`1..3`, `Enter`) unchanged.
- [x] Add tests for item icon resolution fallback (unknown item -> slot icon -> default icon).

### D) Class select improvements (skill preview requirement)
- [x] Extend class select screen to show the 4 class skills with icon + name.
- [x] Ensure selected class highlight and skill preview are visually distinct and readable.
- [x] Keep existing class stat/growth text but rebalance spacing so no overlap occurs.

### E) Buff/debuff tray + target frame
- [x] Implement player buff/debuff tray in run HUD using effect icons + remaining time text.
- [x] Add effect-to-icon mapping for current MVP effects (`Slow`, `Stun`, `Freeze`, `Silence`, `Burn`, `Poison`, plus existing self-buffs used by skills).
- [x] Implement target frame for hovered/locked enemy/boss with HP bar, name/type, elite marker, and active debuff icons.
- [x] Use existing target lock behavior from runtime input/attack logic; do not change targeting rules.

### F) Combat feedback clarity
- [ ] Add floating damage numbers with crit styling (color/size distinction).
- [ ] Add floating heal numbers.
- [ ] Add status popups for major CC states (at least `Stunned`, `Silenced`, `Frozen`).
- [ ] Keep existing telegraphs/hit flash; add settings-driven toggle for small screenshake on impactful hits.

### G) Menus and UX completeness
- [ ] Implement pause menu in `StateRun` flow (`Resume`, `Restart Run`, `Settings`, `Exit to Menu`).
- [ ] Add lightweight settings panel usable from main menu and pause menu (volume/fullscreen + screenshake toggle at minimum).
- [ ] Improve results screen layout to include clearer build summary (class, elapsed time, rooms, picked reward, key stats snapshot).
- [ ] Ensure all menu actions are keyboard-usable and preserve current control conventions.

### H) HUD layout, readability, and polish pass
- [x] Replace text-only HP/resource with readable bars while keeping numeric values.
- [x] Add XP/level/stat-points panel cleanup for hierarchy and spacing.
- [x] Add a simple minimap (room nodes + current room + boss room).
- [x] Ensure HUD layering is stable (world VFX below HUD, overlays above world, debug overlay still works with `F3`).
- [ ] Verify no UI clipping at 1600x900 and at smaller window sizes currently used during development.

### I) Validation checklist (must pass before marking step complete)
- [ ] `go test ./...` passes.
- [ ] Manual smoke: boot -> main menu -> class select -> run -> reward -> results -> main menu with no UI regressions.
- [ ] Manual smoke: each class skill bar shows proper icons and cooldown/cost readability.
- [ ] Manual smoke: reward cards show icons and can be selected with `1..3` + `Enter`.
- [ ] Manual smoke: buff/debuff tray and target frame update correctly during combat.
- [ ] Manual smoke: missing icon texture path still degrades gracefully (fallback visuals, no crash).
