<template>
  <main class="faq-page">
    <div class="faq-container">
      <h1 class="faq-title">Часто задаваемые вопросы</h1>
      <p class="faq-subtitle">Все, что вам нужно знать об игре Сапер Онлайн</p>

      <div class="faq-content">
        <section class="faq-section">
          <h2 class="faq-section-title">
            <IconGamepad class="faq-section-icon" />
            Об игре
          </h2>

          <div class="faq-item">
            <h3 class="faq-question">Что такое Сапер Онлайн?</h3>
            <div class="faq-answer">
              <p>Сапер Онлайн — это многопользовательская версия классической игры Сапер (Minesweeper). Вы можете играть вместе с друзьями в реальном времени, создавать комнаты, соревноваться и получать рейтинг за успешные игры.</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Как играть в Сапера?</h3>
            <div class="faq-answer">
              <p>Цель игры — открыть все ячейки на поле, не наступив на мину. Правила просты:</p>
              <ul>
                <li><strong>Левый клик</strong> — открыть ячейку</li>
                <li><strong>Правый клик</strong> — поставить/убрать флаг на предполагаемую мину</li>
                <li>Числа показывают, сколько мин находится в соседних ячейках</li>
                <li>Используйте логику, чтобы определить расположение мин</li>
              </ul>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Можно ли играть без регистрации?</h3>
            <div class="faq-answer">
              <p>Да! Вы можете играть как гость, просто введя никнейм. Однако для сохранения статистики, получения рейтинга и доступа к профилю рекомендуется зарегистрироваться.</p>
            </div>
          </div>
        </section>

        <section class="faq-section">
          <h2 class="faq-section-title">
            <IconTrophy class="faq-section-icon" />
            Рейтинг и статистика
          </h2>

          <div class="faq-item">
            <h3 class="faq-question">Как работает рейтинговая система?</h3>
            <div class="faq-answer">
              <p>Рейтинг начисляется только за игры на достаточно сложных полях. Чем быстрее вы решите поле и чем сложнее оно было, тем больше рейтинга вы получите.</p>
              <p>Рейтинг учитывает:</p>
              <ul>
                <li>Размер поля (строки × столбцы)</li>
                <li>Количество мин</li>
                <li>Время прохождения</li>
                <li>Ваш текущий рейтинг</li>
              </ul>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Почему за игру не начислили рейтинг?</h3>
            <div class="faq-answer">
              <p>Рейтинг начисляется только за игры на полях достаточной
                сложности. Поля с низкой плотностью мин считаются
                слишком простыми и не дают рейтинга. Для получения рейтинга
                создавайте или присоединяйтесь к комнатам с более сложными настройками.</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Где посмотреть свою статистику?</h3>
            <div class="faq-answer">
              <p>Если вы авторизованы, перейдите в раздел "Профиль" в верхнем меню. Там вы увидите:</p>
              <ul>
                <li>Ваш текущий рейтинг</li>
                <li>Количество сыгранных игр</li>
                <li>Количество побед</li>
                <li>Процент побед</li>
                <li>Топ-100 лучших игр с начисленным рейтингом</li>
              </ul>
            </div>
          </div>
        </section>

        <section class="faq-section">
          <h2 class="faq-section-title">
            <IconUsers class="faq-section-icon" />
            Многопользовательский режим
          </h2>

          <div class="faq-item">
            <h3 class="faq-question">Как создать комнату?</h3>
            <div class="faq-answer">
              <p>На главной странице нажмите кнопку "Создать комнату". Вы сможете настроить:</p>
              <ul>
                <li>Название комнаты</li>
                <li>Размер поля (от 5×5 до 50×50)</li>
                <li>Количество мин</li>
                <li>Пароль для защиты комнаты (опционально)</li>
              </ul>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Как присоединиться к комнате?</h3>
            <div class="faq-answer">
              <p>На главной странице выберите комнату из списка и нажмите на неё. Если комната защищена паролем, вам нужно будет ввести пароль для входа.</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Сколько игроков могут играть одновременно?</h3>
            <div class="faq-answer">
              <p>В одной комнате может играть неограниченное количество игроков одновременно. Все игроки видят действия друг друга в реальном времени и могут общаться в чате.</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Почему я вижу курсоры на поле?</h3>
            <div class="faq-answer">
              <p>Когда другие игроки находятся в той же комнате, вы видите их курсоры на игровом поле. Это помогает понимать, где находятся другие игроки и что они делают.</p>
            </div>
          </div>
        </section>

        <section class="faq-section">
          <h2 class="faq-section-title">
            <IconLightbulb class="faq-section-icon" />
            Подсказки и функции
          </h2>

          <div class="faq-item">
            <h3 class="faq-question">Как работают подсказки?</h3>
            <div class="faq-answer">
              <p>В каждой игре у вас есть 3 подсказки. Подсказка либо открывает случайную
                ячейку, либо ставит флаг на мину. Используйте их с умом, когда застряли!</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Есть ли чат в игре?</h3>
            <div class="faq-answer">
              <p>Да! В каждой комнате есть чат, где вы можете общаться с другими игроками. Чат отображается справа от игрового поля.</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Можно ли играть на мобильных устройствах?</h3>
            <div class="faq-answer">
              <p>Да, игра адаптирована для мобильных устройств. На мобильных есть специальные кнопки зума для удобного управления игровым полем. Для открытия ячейки используйте короткое нажатие, для флага — долгое нажатие.</p>
            </div>
          </div>
        </section>

        <section class="faq-section">
          <h2 class="faq-section-title">
            <IconSettings class="faq-section-icon" />
            Технические вопросы
          </h2>

          <div class="faq-item">
            <h3 class="faq-question">Что делать, если игра не загружается?</h3>
            <div class="faq-answer">
              <p>Попробуйте:</p>
              <ul>
                <li>Обновить страницу (F5 или Ctrl+R)</li>
                <li>Очистить кэш браузера</li>
                <li>Проверить подключение к интернету</li>
                <li>Попробовать другой браузер</li>
              </ul>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Поддерживается ли темная тема?</h3>
            <div class="faq-answer">
              <p>Да! В правом верхнем углу есть переключатель темы. Вы можете выбрать светлую или темную тему в зависимости от ваших предпочтений.</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Сохраняется ли прогресс игры?</h3>
            <div class="faq-answer">
              <p>Если вы авторизованы, ваша статистика и рейтинг сохраняются автоматически. Однако сама игра не сохраняется — если вы выйдете из комнаты, игра начнется заново при следующем входе.</p>
            </div>
          </div>
        </section>

        <section class="faq-section">
          <h2 class="faq-section-title">
            <IconQuestion class="faq-section-icon" />
            Другое
          </h2>

          <div class="faq-item">
            <h3 class="faq-question">Можно ли изменить цвет игрока?</h3>
            <div class="faq-answer">
              <p>Да! В своем профиле вы можете выбрать цвет, который будет отображаться на вашем курсоре и флаге в игре. Доступно множество цветов на выбор.</p>
            </div>
          </div>

          <div class="faq-item">
            <h3 class="faq-question">Как связаться с поддержкой?</h3>
            <div class="faq-answer">
              <p>Если у вас возникли проблемы или вопросы, которые не освещены в этом FAQ, пожалуйста, сообщите об этом на почту разработчика:</p>
              <div class="support-email">
                <a href="mailto:guhbap@gmail.com" class="support-email-link">
                  <IconEmail class="email-icon" />
                  <span class="email-address">guhbap@gmail.com</span>
                </a>
              </div>
            </div>
          </div>
        </section>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import IconGamepad from '@/components/icons/IconGamepad.vue'
import IconTrophy from '@/components/icons/IconTrophy.vue'
import IconUsers from '@/components/icons/IconUsers.vue'
import IconLightbulb from '@/components/icons/IconLightbulb.vue'
import IconSettings from '@/components/icons/IconSettings.vue'
import IconQuestion from '@/components/icons/IconQuestion.vue'
import IconEmail from '@/components/icons/IconEmail.vue'
</script>

<style scoped>
.faq-page {
  min-height: 100vh;
  padding: 2rem;
  background: var(--bg-primary);
}

.faq-container {
  max-width: 900px;
  margin: 0 auto;
}

.faq-title {
  font-size: 2.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 0.5rem 0;
  text-align: center;
}

.faq-subtitle {
  font-size: 1.125rem;
  color: var(--text-secondary);
  text-align: center;
  margin: 0 0 3rem 0;
}

.faq-content {
  display: flex;
  flex-direction: column;
  gap: 2.5rem;
}

.faq-section {
  background: var(--bg-secondary);
  border-radius: 1rem;
  padding: 2rem;
  box-shadow: 0 2px 8px var(--shadow);
}

.faq-section-title {
  font-size: 1.5rem;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0 0 1.5rem 0;
  padding-bottom: 0.75rem;
  border-bottom: 2px solid var(--border-color);
}

.faq-item {
  margin-bottom: 1.5rem;
}

.faq-item:last-child {
  margin-bottom: 0;
}

.faq-question {
  font-size: 1.125rem;
  font-weight: 600;
  color: var(--text-primary);
  margin: 0 0 0.75rem 0;
}

.faq-answer {
  color: var(--text-secondary);
  line-height: 1.6;
}

.faq-answer p {
  margin: 0 0 0.75rem 0;
}

.faq-answer p:last-child {
  margin-bottom: 0;
}

.faq-answer ul {
  margin: 0.75rem 0;
  padding-left: 1.5rem;
}

.faq-answer li {
  margin-bottom: 0.5rem;
}

.faq-answer li:last-child {
  margin-bottom: 0;
}

.faq-answer strong {
  color: var(--text-primary);
  font-weight: 600;
}

.support-email {
  margin: 1rem 0;
  display: flex;
  justify-content: center;
  align-items: center;
}

.support-email-link {
  display: inline-flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0.875rem 1.5rem;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #ffffff;
  text-decoration: none;
  border-radius: 0.75rem;
  font-weight: 600;
  font-size: 1rem;
  transition: all 0.2s ease-in-out;
  box-shadow: 0 2px 8px rgba(102, 126, 234, 0.3);
}

.support-email-link:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 12px rgba(102, 126, 234, 0.4);
  text-decoration: none;
  color: #ffffff;
}

.support-email-link:active {
  transform: translateY(0);
  opacity: 0.9;
}

.email-icon {
  width: 1.25rem;
  height: 1.25rem;
  flex-shrink: 0;
}

.faq-section-icon {
  width: 1.5rem;
  height: 1.5rem;
  display: inline-block;
  vertical-align: middle;
  margin-right: 0.5rem;
}

.email-address {
  font-family: 'Courier New', monospace;
  letter-spacing: 0.5px;
}

@media (max-width: 768px) {
  .faq-page {
    padding: 1rem;
  }

  .faq-title {
    font-size: 2rem;
  }

  .faq-subtitle {
    font-size: 1rem;
    margin-bottom: 2rem;
  }

  .faq-section {
    padding: 1.5rem;
  }

  .faq-section-title {
    font-size: 1.25rem;
  }

  .faq-question {
    font-size: 1rem;
  }
}
</style>

