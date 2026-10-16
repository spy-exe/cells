# Cells

Um jogo multiplayer web de células, com simulação autoritativa, Classic FFA, bots táticos e laboratório de IA. Projeto open source de spy-exe, licenciado sob AGPL-3.0. Marca, interface e conteúdo próprios.

Destino: https://cells.malha.app. O desenvolvimento está na fundação; os gates de produção serão registrados em `CONTINUITY.md`, sem apresentar modos ainda não implementados como jogáveis.

[Contrato da entrega](SPEC.md) · [Infraestrutura descoberta](docs/INFRASTRUCTURE.md) · [Arquitetura](docs/adr/0001-modular-go-server.md)

A simulação ficará em Go, reutilizada pelo servidor e pelo ambiente headless; o cliente React/TypeScript/PixiJS renderiza e envia intenção. PostgreSQL guarda dados persistentes. Treino e inferência são medidos separadamente; bots heurísticos não são apresentados como modelos treinados.

Os comandos e resultados verificáveis acompanham cada release. Não coloque credenciais, dados reais de sessões ou backups neste repositório público.
