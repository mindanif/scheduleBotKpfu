package bot

import (
	"context"
	"log"
	"strconv"
	"strings"

	tg "github.com/go-telegram/bot"
	modelsTg "github.com/go-telegram/bot/models"
	//"scheduleBot/formatter"
	"scheduleBot/models"
	"scheduleBot/repository"
	"scheduleBot/schedule"
)

type TelegramBot struct {
	bot              *tg.Bot
	userRepo         repository.UserRepository
	teacherRepo      repository.TeacherRepository
	scheduleProvider schedule.ScheduleProvider
	//scheduleFormatter formatter.ScheduleFormatter
}

func NewTelegramBot(
	token string,
	userRepo repository.UserRepository,
	teacherRepo repository.TeacherRepository,
	scheduleProvider schedule.ScheduleProvider,
) (*TelegramBot, error) {

	botAPI, err := tg.New(token)
	if err != nil {
		return nil, err
	}

	return &TelegramBot{
		bot:              botAPI,
		userRepo:         userRepo,
		teacherRepo:      teacherRepo,
		scheduleProvider: scheduleProvider,
	}, nil
}

func (t *TelegramBot) Start() {
	me, err := t.bot.GetMe(context.Background())
	if err != nil {
		log.Fatalf("Ошибка getMe: %v", err)
	}
	log.Printf("Бот запущен как: @%s", me.Username)

	// Команды
	t.bot.RegisterHandler(tg.HandlerTypeMessageText, "/start", tg.MatchTypeExact, t.onStart)
	t.bot.RegisterHandler(tg.HandlerTypeMessageText, "/reset", tg.MatchTypeExact, t.onReset)
	t.bot.RegisterHandler(tg.HandlerTypeMessageText, "/web", tg.MatchTypeExact, t.onWeb)
	t.bot.RegisterHandler(tg.HandlerTypeMessageText, "", tg.MatchTypePrefix, t.onText)
	t.bot.RegisterHandler(tg.HandlerTypeCallbackQueryData, "register:", tg.MatchTypePrefix, t.onRegisterCallback)

	t.bot.Start(context.Background())
}
func (t *TelegramBot) onStart(ctx context.Context, b *tg.Bot, update *modelsTg.Update) {
	chatID := update.Message.Chat.ID
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: chatID,
		Text:   "Добро пожаловать! Введите ФИО преподавателя для регистрации:",
	})
}
func (t *TelegramBot) onReset(ctx context.Context, b *tg.Bot, update *modelsTg.Update) {
	chatID := update.Message.Chat.ID
	if _, err := t.userRepo.Delete(chatID); err != nil {
		return
	}
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: chatID,
		Text:   "Регистрация сброшена! Введите ФИО заново:",
	})
}
func (t *TelegramBot) onWeb(ctx context.Context, b *tg.Bot, update *modelsTg.Update) {
	chatID := update.Message.Chat.ID
	user, err := t.userRepo.Get(chatID)
	if err != nil || !user.Registered {
		t.sendRegistrationPrompt(ctx, b, chatID)
		return
	}
	t.OpenWebApp(ctx, b, chatID)
}
func (t *TelegramBot) sendRegistrationPrompt(ctx context.Context, b *tg.Bot, chatID int64) {
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: chatID,
		Text:   "Пожалуйста, введите ФИО преподавателя для регистрации:",
	})
}

// Не работает
func (t *TelegramBot) onWebAppData(ctx context.Context, b *tg.Bot, update *modelsTg.Update) {
	if update.Message != nil && update.Message.WebAppData != nil {
		data := update.Message.WebAppData.Data // это строка, которую отправил Web App
		log.Printf("Получены данные из Web App: %s", data)

		b.SendMessage(ctx, &tg.SendMessageParams{
			ChatID: update.Message.Chat.ID,
			Text:   "Данные из Web App получены: " + data,
		})
	}
}
func (t *TelegramBot) onText(ctx context.Context, b *tg.Bot, update *modelsTg.Update) {
	if update.Message != nil && update.Message.WebAppData != nil {
		t.onWebAppData(ctx, b, update)
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text

	user, err := t.userRepo.Get(chatID)
	if err != nil || !user.Registered {
		t.handleRegistrationWeb(ctx, b, chatID, text)
		return
	}

	switch text {
	case "Расписание на сегодня":
		t.sendScheduleWeb(ctx, b, chatID, user.SelectedTeacher.ID, "сегодня")
	default:
		if day := parseDay(text); day != "" {
			t.sendScheduleWeb(ctx, b, chatID, user.SelectedTeacher.ID, day)
		} else {
			b.SendMessage(ctx, &tg.SendMessageParams{
				ChatID: chatID,
				Text:   "Неизвестная команда. Выберите опцию из меню.",
			})
		}
	}
}
func (t *TelegramBot) onRegisterCallback(ctx context.Context, b *tg.Bot, update *modelsTg.Update) {
	cq := update.CallbackQuery
	chatID := cq.From.ID
	teacherID := strings.TrimPrefix(cq.Data, "register:")

	teacher, err := t.teacherRepo.GetByID(teacherID)
	if err != nil {
		b.AnswerCallbackQuery(ctx, &tg.AnswerCallbackQueryParams{
			CallbackQueryID: cq.ID,
			Text:            "Преподаватель не найден.",
		})
		return
	}
	t.userRepo.Save(models.User{
		ChatID:          chatID,
		SelectedTeacher: teacher,
		Registered:      true,
	})
	b.AnswerCallbackQuery(ctx, &tg.AnswerCallbackQueryParams{
		CallbackQueryID: cq.ID,
		Text:            "Регистрация успешна!",
	})
	t.sendMainMenuWeb(ctx, b, chatID)
}
func (t *TelegramBot) sendMainMenuWeb(ctx context.Context, b *tg.Bot, chatID int64) {
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: chatID,
		Text:   "Выберите опцию расписания:",
		ReplyMarkup: &modelsTg.ReplyKeyboardMarkup{
			ResizeKeyboard: true,
			Keyboard: [][]modelsTg.KeyboardButton{
				{
					{Text: "Расписание на сегодня"},
					{Text: "Понедельник"},
				},
				{
					{Text: "Вторник"},
					{Text: "Среда"},
				},
				{
					{Text: "Четверг"},
					{Text: "Пятница"},
				},
				{
					{Text: "Суббота"},
					{Text: "Воскресенье"},
				},
			},
		},
	})
}
func (t *TelegramBot) handleRegistrationWeb(ctx context.Context, b *tg.Bot, chatID int64, fullName string) {
	matches, err := t.teacherRepo.FindByName(fullName)
	if err != nil || len(matches) == 0 {
		b.SendMessage(ctx, &tg.SendMessageParams{
			ChatID: chatID,
			Text:   "Не найдено совпадений. Попробуйте ввести ФИО еще раз.",
		})
		return
	}

	messageText := "Найдено совпадений:\n"
	var rows [][]modelsTg.InlineKeyboardButton
	for i, teacher := range matches {
		messageText += strconv.Itoa(i+1) + ". " + teacher.FullName + "\n"
		if i%8 == 0 {
			rows = append(rows, nil)
		}
		rows[len(rows)-1] = append(rows[len(rows)-1], modelsTg.InlineKeyboardButton{
			Text:         strconv.Itoa(i + 1),
			CallbackData: "register:" + teacher.ID,
		})
	}
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID: chatID,
		Text:   messageText + "\nВыберите, пожалуйста, номер, соответствующий вашему ФИО:",
		ReplyMarkup: &modelsTg.InlineKeyboardMarkup{
			InlineKeyboard: rows,
		},
	})
}
func (t *TelegramBot) sendScheduleWeb(ctx context.Context, b *tg.Bot, chatID int64, teacherID, day string) {
	sched, err := t.scheduleProvider.GetSchedule(teacherID)
	if err != nil {
		b.SendMessage(ctx, &tg.SendMessageParams{
			ChatID: chatID,
			Text:   "Ошибка при получении расписания.",
		})
		return
	}
	text := sched.FormatForDay(day)
	b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: "Markdown",
	})
}
func (t *TelegramBot) OpenWebApp(ctx context.Context, b *tg.Bot, chatID int64) {
	user, err := t.userRepo.Get(chatID)
	if err != nil || !user.Registered {
		return
	}
	webAppURL := "https://schedulebot-production.up.railway.app/web" + "?id=" + user.SelectedTeacher.ID
	inlineKeyboard := &modelsTg.InlineKeyboardMarkup{
		InlineKeyboard: [][]modelsTg.InlineKeyboardButton{
			{
				{
					Text: "Открыть веб-приложение",
					WebApp: &modelsTg.WebAppInfo{
						URL: webAppURL,
					},
				},
			},
		},
	}
	_, err = b.SendMessage(ctx, &tg.SendMessageParams{
		ChatID:      chatID,
		Text:        "Нажмите кнопку ниже для открытия веб‑приложения:",
		ReplyMarkup: inlineKeyboard,
	})
	if err != nil {
		log.Println(err)
	}
}

func parseDay(text string) string {
	dayNames := []string{
		"Понедельник", "Вторник", "Среда",
		"Четверг", "Пятница", "Суббота",
		"Воскресенье", "сегодня",
	}
	for _, day := range dayNames {
		if strings.EqualFold(text, day) {
			return day
		}
	}
	return ""
}
