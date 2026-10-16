# ADR 0004 — treinamento real em laboratório isolado

Estado: implementação em andamento, após Cells 0.2.2/c11026d publicado e verificado.
Nenhum modelo foi treinado ainda. Produção usa políticas heurísticas identificadas.

O binário Go headless reutiliza engine/bot e não abre rede nem banco. Stdio JSON
v1 oferece reset/step/close, ações simultâneas para 1–4 agentes e o mesmo Tick30Hz.
Cada decisão avança seis ticks. Episódios têm horizonte finito, sem respawn, com
hash final e streams de seed/ações para replay. Validação acontece antes de mutar
o mundo. O cliente de treinamento recebe observações locais e métricas próprias,
sem RNG, entidades invisíveis ou cooldowns de inimigos.

Arena inicial de treino: 3200 unidades, 1200 alimentos e 12 vírus. Coeficientes,
consumo, split/eject/merge e contabilidade de massa são os mesmos da produção.
A arena compacta facilita encontros com orçamento pequeno; sua avaliação não
autoriza automaticamente promoção ao Classic de 6000 unidades.

Observação: vetor de 80 floats em [-1,1]. Doze atributos próprios/geométricos,
seis alimentos próximos (dx,dy,massa), oito células alheias próximas (dx,dy,razão
de massa,ameaça,aliado), três vírus próximos (dx,dy,raio) e densidade local de
alimento. Features derivam somente de engine.Observation. Ação Box4: direçãoXY
normalizada à visão, split se componente2>0.5, eject se componente3>0.5.

Reward: delta log(1+massa), +0.2 por eliminação, -1 ao morrer, -0.001 por decisão,
e bônus únicos0.05 ao sobreviver100/300/500decisões. O saldo passivo é negativo;
ejetar e recolher a própria massa tem saldo negativo. Testes devem impedir esses
exploits. Métricas são separadas do reward: massa final/pico/AUC, sobrevivência,
eliminações, vitória por massa final e custo real da inferência.

Python: Gymnasium para SB3 PPO e PettingZoo ParallelEnv para múltiplos agentes.
Pinagens confirmadas no PyPI: SB3 2.9.0, Gymnasium1.4.0, PettingZoo1.27.0,
NumPy2.5.4; PyTorch2.14.1+cpu confirmado no índice CPU oficial (cp313manylinux).
Instalados no CT221; validar imports/pipcheck antes de executar. Dependências
transitivas e hashes do modelo serão registrados a partir da execução real.

Plano do primeiro experimento: PPO MLP64x64, CPU/thread1, n_steps512,batch64,
n_epochs4,gamma0.995,lr0.0003,32768passos, curriculum0/3/7oponentes. Treino e
avaliação usam CT221/cap0.75core; unidade até65%CPU e ~1.1GiB, com timeout.
Produção CT220 não executa treino. Salvamentos periódicos, retomada efetivamente
testada, manifesto de versão/seed/dependências/checksum. Treino usa seeds<100000;
avaliação usa100001..100100 por política random/heurística/PPO, mesma configuração.
Guardar resultados por episódio, poucos replays e gráficos derivados de dados.
Medir custo antes de ampliar o orçamento; não apresentar metas como resultados.

Promoção exige comparação pareada, ganho relevante sobre a heurística, segurança,
latência e validação na configuração de produção. Se PPO perder, manter heurística
em produção e publicar o resultado. Arquivos de modelo são confiáveis somente
quando gerados pelo pipeline e verificados; não carregar pickle de usuários.

Referências primárias consultadas:
- https://stable-baselines3.readthedocs.io/en/master/guide/custom_env.html
- https://pettingzoo.farama.org/api/parallel/
- https://pytorch.org/get-started/locally/
- https://download.pytorch.org/whl/cpu/torch/

Após a base: melhorar coerência das heurísticas com memória de objetivos/rotas,
comparar versões, acrescentar feedback visual e resumo de rodada com estatísticas
autoritativas. Atualizar clone de bot.Memory nos checkpoints ao adicionar campos.
Nenhum modo ou medidor de treinamento fictício será exposto na interface.
