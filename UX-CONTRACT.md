# Contrato de interação — Cells

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Button | style.css button/play-button | UX-CONTRACT.md | foco, hover, disabled com motivo, busy estável | tests/e2e/foundation.spec.ts |
| Form | entrada main.tsx e API Go | UX-CONTRACT.md | label, noValidate, nickname limitado; erros inline | tests/e2e/foundation.spec.ts |
| Scrollbar | style.css :root e fallbacks | DESIGN.md | herdado por scrollers, forced colors | tests/e2e/foundation.spec.ts |
| Navigation | links nativos e estado React | UX-CONTRACT.md | entrada e fonte; abas futuras acessíveis | tests/e2e/foundation.spec.ts |
| Motion | data-motion e prefers-reduced-motion | DESIGN.md | preferência persistente e sistema | tests/e2e/foundation.spec.ts |

Status é obtido da API e validado em packages/protocol. Só modos ativos são anunciados. A fundação identifica o jogo em construção, sem botão funcional fictício. Falha de status é recuperável e não fabrica presença. Nickname e preferências não são credenciais; sessões futuras usam cookie HttpOnly no servidor.
