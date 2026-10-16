# Contrato de interação — Cells

Produto em português, com entrada rápida e arena como protagonista. Os controles alteram apenas intenções; massa, velocidade e cooldowns pertencem ao servidor.

| Capability | Canonical owner | Source of truth | Allowed variants | Verification |
|---|---|---|---|---|
| Button | style.css button/play-button | UX-CONTRACT.md | foco, hover, disabled com motivo, busy estável | tests/e2e/foundation.spec.ts |
| Form | entrada main.tsx e API Go | UX-CONTRACT.md | label, noValidate, nickname limitado, erro inline | tests/e2e/foundation.spec.ts |
| Scrollbar | style.css :root e fallbacks | DESIGN.md | herdado por scrollers, forced colors | tests/e2e/foundation.spec.ts |
| Navigation | links nativos e estado React | UX-CONTRACT.md | entrada, partida, saída e código fonte | tests/e2e/classic.spec.ts |
| Motion | data-motion e prefers-reduced-motion | DESIGN.md | preferência persistente e sistema | tests/e2e/foundation.spec.ts |
| Select/Listbox | native | UX-CONTRACT.md | seletores nativos de modo e qualidade, label, foco e teclado | tests/e2e/classic.spec.ts |
| Arena controls | Game.tsx e engine.Input | docs/architecture/protocol.md | mouse, setas, toque e direcional; split/eject compartilham cooldowns | tests/e2e/classic.spec.ts |

O navegador controla teclado, picker de toque e semântica dos selects. style.css é o proprietário visual comum. Só modos ativos anunciados pela API aparecem no menu. Falhas de status são recuperáveis e não fabricam presença. Sessões usam cookie HttpOnly no servidor.

A simulação continua durante menus e desconexões. A morte muda a câmera para o líder e permite respawn após o prazo do servidor. A reconexão preserva a identidade enquanto o corpo continua na arena por até 30 segundos. Outra janela com a mesma sessão substitui a conexão anterior, com mensagem explícita. Preferências de movimento afetam efeitos decorativos; a opção econômica limita a renderização a 30 FPS, sem alterar os ticks do servidor.

Ações são enviadas no pressionamento. Segurar W ou o botão Ejetar mantém a ejeção enquanto há massa suficiente; soltar, cancelar o toque ou perder foco interrompe a repetição. A câmera e as células usam a mesma linha de tempo interpolada. Surgimento, crescimento, desaparecimento e feedback de botões respeitam movimento reduzido. Células usam geometria vetorial para manter o contorno nítido com zoom e crescimento; a forma dos vírus tem 48 pontas.

Disponibilidade das ações considera a massa individual: pelo menos uma célula com massa 36 para dividir ou 32 para ejetar, além do cooldown e do limite de 16 células. Quando falta massa, o botão mostra o requisito por célula. A massa total do jogador não substitui essa condição.
