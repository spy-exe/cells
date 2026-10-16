# ADR 0001 — servidor modular e simulação compartilhada

Aceito em 16/10/2026. Um executável Go serve API, assets, WebSocket e métricas privadas; salas têm loops autoritativos de tick fixo. Regras em pacote independente da rede e da UI, reutilizado pelo executável headless de treinamento. PostgreSQL é local. Não usar Docker, Kubernetes ou serviços distribuídos no primeiro lançamento.

O cliente recebe snapshots binários por interesse e interpola; só envia intenção de movimento/ações. Bots usam a observação filtrada. Treino ocorre em ambiente separado do processo/CT de jogo, com orçamento de CPU/memória, e não muda pesos de produção automaticamente. Modes implementam política de spawn/elegibilidade de consumo/encerramento sobre a mesma física.

Isso reduz operações e evita divergência entre treino e jogo. Limite inicial: um servidor no homelab, sem alta disponibilidade; restauração e rollback testados, cópia externa de dados depende de destino disponível.
