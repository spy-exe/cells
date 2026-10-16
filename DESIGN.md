# Cells — identidade e interface

Produto de jogo; idioma pt-BR. A arena ocupa a tela e recebe o primeiro olhar. O menu é uma camada de entrada rápida, sem aparência de dashboard. Mundo claro em papel mineral, grade pontilhada discreta e células circulares com cores próprias. Marca tipográfica `cells.` e dois círculos sobrepostos, desenhados para este projeto.

Tokens de runtime em apps/web/src/style.css: campo #f4f7ef, tinta #17251c, texto secundário #536456, lima #d1f864, linhas #d4ddcd, painel #ffffff, ação secundária #546f32. Space Grotesk para marca/títulos; Inter para controles e leitura; fontes locais sob OFL. Ícones Lucide. Cores de células são dados visuais e não indicam habilidades.

Mobile: coluna única e touch targets ≥44px, safe areas, ações de dividir/ejetar acessíveis. Desktop: entrada à esquerda, formulário à direita; HUD pequeno e arena plena. Câmera e interpolação não alteram a física. Controles de qualidade, som e movimento têm significado próprio e feedback.

Movimento reduzido do sistema e opção persistente desativam animação decorativa; não mascarar tempo do jogo. Som opcional. Foco visível, formulários validados, erros em texto. Usar diálogo próprio para confirmação sensível. Nenhum gesto será a única forma de agir. O estado inicial nunca anuncia partidas, modos, jogadores ou treinamento inexistentes.

Arte gerada por formas vetoriais originais e renderização em PixiJS. Não usar assets, marca ou interface proprietária de outro jogo. A experiência é inspirada no gênero clássico de células.
