# Cells — contrato de entrega

Jogo web original de células: entrada rápida, Classic FFA sem poderes, Go autoritativo a 30 Hz, React/TypeScript/Vite/PixiJS com interpolação, PostgreSQL local. Interface pt-BR; protocolo binário versionado; identidade própria. Meta de renderização 60 FPS, medida separadamente do tick.

| Módulo | Responsabilidade | Depende de |
|---|---|---|
| engine | física, massa, colisões, split/eject/vírus/merge, seeds, visibilidade | — |
| protocol | comandos e snapshots binários limitados | engine |
| server | salas, sessões, WebSocket, persistência, segurança, métricas | engine, protocol |
| web | renderização, entrada, HUD, espectador, reconexão | protocol, server |
| tactical-ai | percepção equivalente à humana e cinco perfis | engine |
| training | simulador headless, ambiente multiagente, PPO, comparação | engine, tactical-ai |
| community | sugestões moderadas, votos, roadmap, changelog/status | server |
| delivery | LXC, TLS, systemd, backup/restore, CI/CD | server, web |

Ordem: fundação publicada → Classic jogável → bots/AI Arena → treino e avaliação → Teams/comunidade. Cada gate exige execução real. Não expor modos ainda não funcionais.

Aceite do motor: velocidade decresce com massa, limites, massa explícita com fontes/sumidouros, razão/tolerância de consumo, no máximo 16 células, split e eject com custos/cooldowns, vírus explodem célula grande, merge com prazo, spawn seguro, replays por seed. Comandos nunca trazem massa, velocidade ou posição autoritativa.

Aceite de rede: origem validada, mensagens binárias limitadas e versionadas, frequência por conexão/IP, cookie de sessão HttpOnly/Secure, identidade própria, bounded queues, reconexão sem duplicar jogador. API valida entradas/CSRF por origem e JSON; banco não público. Interest management no servidor; bots usam a mesma observação visível.

Aceite visual: mouse/teclado/toque; 360×800, 390×844, 412×915, 768×1024, 1440×900; sem overflow, foco/contraste, movimento reduzido. Dois clientes na mesma partida pública, morte/respawn, split/eject, console/network sem erros inesperados. Pixels são desenhados por Pixi; simulação não mora no navegador.

Gates: Go tests/race/coverage ≥85% engine; TS typecheck/unit/build/audit; Python tests/env checker; Playwright desktop/mobile; cargas 25/50/100 com p95/p99 e recursos; 100 episódios por política quando viável. Não reduzir testes para esconder falhas.

IA: regras heurísticas identificadas; treino PPO real fora do serviço de jogo, com orçamento explícito, seed, modelo e métricas versionados. Comparar random/heurística/treinada com seeds iguais. Só promover modelo que passe critérios publicados. Baseline fica ativo se o modelo não superar. Comunidade não altera automaticamente segurança ou arquitetura.

Comandos e detalhes de regras serão registrados junto com a implementação real. Não criar diretórios vazios ou indicadores fictícios. Documentação final lista o que foi executado, bloqueios e continuidade.
