# ADR 0003 — Explicit WebGL renderer under CSP

Accepted, 2026-10-16.

PixiJS renders cached cell textures and labels in a scene separate from simulation.
An explicit WebGLRenderer, ResizeObserver and bounded requestAnimationFrame loop
support high and economical quality. Browser selection is explicit; unsupported
WebGL displays an actionable error rather than a blank arena.

The pixi.js/unsafe-eval compatibility module supplies dispatch polyfills that avoid
eval. Its name does not mean relaxing CSP. The shipped CSP still excludes unsafe-eval.
See the installed PixiJS module and the official migration guide:
https://pixijs.com/8.x/guides/migrations/v8

A visual QA failure showed that connected HUDs are insufficient evidence of a game.
The E2E gate now requires the actual canvas and measured render frames, and real
screenshots are inspected. Render initialization errors cannot be erased by a
successful WebSocket connection.
