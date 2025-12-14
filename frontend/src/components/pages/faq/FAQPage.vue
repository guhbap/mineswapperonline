<template>
  <main class="faq-page">
    <div class="faq-container">
      <h1 class="faq-title">Часто задаваемые вопросы</h1>
      <p class="faq-subtitle">Все, что вам нужно знать об игре Сапер Онлайн</p>

      <div class="faq-content">
        <section class="faq-section">
          <h2 class="faq-section-title">
            <svg class="faq-section-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <rect x="2" y="6" width="20" height="12" rx="2" stroke="currentColor" stroke-width="2"/>
              <path d="M6 10H18M6 14H14" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            </svg>
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
            <svg class="faq-section-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M6 9H4.5C3.67157 9 3 9.67157 3 10.5V19.5C3 20.3284 3.67157 21 4.5 21H19.5C20.3284 21 21 20.3284 21 19.5V10.5C21 9.67157 20.3284 9 19.5 9H18M6 9V6C6 4.34315 7.34315 3 9 3H15C16.6569 3 18 4.34315 18 6V9M6 9H18M12 9V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M9 12L12 9L15 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
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
            <svg class="faq-section-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M17 21V19C17 17.9391 16.5786 16.9217 15.8284 16.1716C15.0783 15.4214 14.0609 15 13 15H5C3.93913 15 2.92172 15.4214 2.17157 16.1716C1.42143 16.9217 1 17.9391 1 19V21" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <circle cx="9" cy="7" r="4" stroke="currentColor" stroke-width="2"/>
              <path d="M23 21V19C22.9993 18.1137 22.7044 17.2528 22.1614 16.5523C21.6184 15.8519 20.8581 15.3516 20 15.13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M16 3.13C16.8604 3.35031 17.623 3.85071 18.1676 4.55232C18.7122 5.25392 19.0078 6.11683 19.0078 7.005C19.0078 7.89318 18.7122 8.75608 18.1676 9.45769C17.623 10.1593 16.8604 10.6597 16 10.88" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
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
            <svg class="faq-section-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 2V4M12 20V22M4 12H2M6.31412 6.31412L4.8999 4.8999M17.6859 6.31412L19.1001 4.8999M6.31412 17.69L4.8999 19.1042M17.6859 17.69L19.1001 19.1042M22 12H20M17 12C17 14.7614 14.7614 17 12 17C9.23858 17 7 14.7614 7 12C7 9.23858 9.23858 7 12 7C14.7614 7 17 9.23858 17 12Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
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
            <svg class="faq-section-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 15C13.6569 15 15 13.6569 15 12C15 10.3431 13.6569 9 12 9C10.3431 9 9 10.3431 9 12C9 13.6569 10.3431 15 12 15Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M19.4 15C19.2669 15.3016 19.2272 15.6362 19.286 15.9606C19.3448 16.285 19.4995 16.5843 19.73 16.82L19.79 16.88C19.976 17.0657 20.1235 17.2863 20.2241 17.5291C20.3248 17.7719 20.3766 18.0322 20.3766 18.295C20.3766 18.5578 20.3248 18.8181 20.2241 19.0609C20.1235 19.3037 19.976 19.5243 19.79 19.71C19.6043 19.896 19.3837 20.0435 19.1409 20.1441C18.8981 20.2448 18.6378 20.2966 18.375 20.2966C18.1122 20.2966 17.8519 20.2448 17.6091 20.1441C17.3663 20.0435 17.1457 19.896 16.96 19.71L16.9 19.65C16.6643 19.4195 16.365 19.2648 16.0406 19.206C15.7162 19.1472 15.3816 19.1869 15.08 19.32C14.7842 19.4468 14.532 19.6572 14.3543 19.9255C14.1766 20.1938 14.0813 20.5082 14.08 20.83V21C14.08 21.5304 13.8693 22.0391 13.4942 22.4142C13.1191 22.7893 12.6104 23 12.08 23C11.5496 23 11.0409 22.7893 10.6658 22.4142C10.2907 22.0391 10.08 21.5304 10.08 21V20.91C10.0723 20.579 9.96512 20.258 9.77251 19.9887C9.5799 19.7194 9.31074 19.5143 9 19.4C8.69838 19.2669 8.36381 19.2272 8.03941 19.286C7.71502 19.3448 7.41568 19.4995 7.18 19.73L7.12 19.79C6.93425 19.976 6.71368 20.1235 6.47088 20.2241C6.22808 20.3248 5.96783 20.3766 5.705 20.3766C5.44217 20.3766 5.18192 20.3248 4.93912 20.2241C4.69632 20.1235 4.47575 19.976 4.29 19.79C4.10405 19.6043 3.95653 19.3837 3.85588 19.1409C3.75523 18.8981 3.70343 18.6378 3.70343 18.375C3.70343 18.1122 3.75523 17.8519 3.85588 17.6091C3.95653 17.3663 4.10405 17.1457 4.29 16.96L4.35 16.9C4.58054 16.6643 4.73519 16.365 4.794 16.0406C4.85282 15.7162 4.81312 15.3816 4.68 15.08C4.55324 14.7842 4.34276 14.532 4.07447 14.3543C3.80618 14.1766 3.49179 14.0813 3.17 14.08H3C2.46957 14.08 1.96086 13.8693 1.58579 13.4942C1.21071 13.1191 1 12.6104 1 12.08C1 11.5496 1.21071 11.0409 1.58579 10.6658C1.96086 10.2907 2.46957 10.08 3 10.08H3.09C3.42099 10.0723 3.742 9.96512 4.01131 9.77251C4.28062 9.5799 4.48571 9.31074 4.6 9C4.73312 8.69838 4.77282 8.36381 4.714 8.03941C4.65519 7.71502 4.50054 7.41568 4.27 7.18L4.21 7.12C4.02405 6.93425 3.87653 6.71368 3.77588 6.47088C3.67523 6.22808 3.62343 5.96783 3.62343 5.705C3.62343 5.44217 3.67523 5.18192 3.77588 4.93912C3.87653 4.69632 4.02405 4.47575 4.21 4.29C4.39575 4.10405 4.61632 3.95653 4.85912 3.85588C5.10192 3.75523 5.36217 3.70343 5.625 3.70343C5.88783 3.70343 6.14808 3.75523 6.39088 3.85588C6.63368 3.95653 6.85425 4.10405 7.04 4.29L7.1 4.35C7.33568 4.58054 7.63502 4.73519 7.95941 4.794C8.28381 4.85282 8.61838 4.81312 8.92 4.68H9C9.29577 4.55324 9.54802 4.34276 9.72569 4.07447C9.90337 3.80618 9.99872 3.49179 10 3.17V3C10 2.46957 10.2107 1.96086 10.5858 1.58579C10.9609 1.21071 11.4696 1 12 1C12.5304 1 13.0391 1.21071 13.4142 1.58579C13.7893 1.96086 14 2.46957 14 3V3.09C14.0013 3.41179 14.0966 3.72618 14.2743 3.99447C14.452 4.26276 14.7042 4.47324 15 4.6C15.3016 4.73312 15.6362 4.77282 15.9606 4.714C16.285 4.65519 16.5843 4.50054 16.82 4.27L16.88 4.21C17.0657 4.02405 17.2863 3.87653 17.5291 3.77588C17.7719 3.67523 18.0322 3.62343 18.295 3.62343C18.5578 3.62343 18.8181 3.67523 19.0609 3.77588C19.3037 3.87653 19.5243 4.02405 19.71 4.21C19.896 4.39575 20.0435 4.61632 20.1441 4.85912C20.2448 5.10192 20.2966 5.36217 20.2966 5.625C20.2966 5.88783 20.2448 6.14808 20.1441 6.39088C20.0435 6.63368 19.896 6.85425 19.71 7.04L19.65 7.1C19.4195 7.33568 19.2648 7.63502 19.206 7.95941C19.1472 8.28381 19.1869 8.61838 19.32 8.92V9C19.4468 9.29577 19.6572 9.54802 19.9255 9.72569C20.1938 9.90337 20.5082 9.99872 20.83 10H21C21.5304 10 22.0391 10.2107 22.4142 10.5858C22.7893 10.9609 23 11.4696 23 12C23 12.5304 22.7893 13.0391 22.4142 13.4142C22.0391 13.7893 21.5304 14 21 14H20.91C20.5882 14.0013 20.2738 14.0966 20.0055 14.2743C19.7372 14.452 19.5268 14.7042 19.4 15Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
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
            <svg class="faq-section-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="2"/>
              <path d="M9.09 9C9.3251 8.33167 9.78915 7.76811 10.4 7.40913C11.0108 7.05016 11.7289 6.91894 12.4272 7.03871C13.1255 7.15849 13.7588 7.52152 14.2151 8.06353C14.6713 8.60553 14.9211 9.29152 14.92 10C14.92 12 11.92 13 11.92 13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
              <path d="M12 17H12.01" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
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
                  <svg class="email-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path d="M4 4H20C21.1 4 22 4.9 22 6V18C22 19.1 21.1 20 20 20H4C2.9 20 2 19.1 2 18V6C2 4.9 2.9 4 4 4Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                    <path d="M22 6L12 13L2 6" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
                  </svg>
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
// Страница FAQ не требует логики
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

