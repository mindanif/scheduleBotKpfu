package main

import (
	"log"
	"scheduleBot/web"

	"scheduleBot/bot"
	"scheduleBot/config"
	"scheduleBot/repository"
	"scheduleBot/schedule"
)

func main() {
	cfg, err := config.NewConfigFromFile("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}

	userRepo := repository.NewInMemoryUserRepository()
	teacherRepo := repository.NewKfuAPITeacherRepository(cfg.KFUApiBaseUrl)

	scheduleProvider := schedule.NewKFUProvider(cfg.KFUApiBaseUrl)

	telegramBot, err := bot.NewTelegramBot(
		cfg.TelegramToken,
		userRepo,
		teacherRepo,
		scheduleProvider,
		//scheduleFormatter,
	)
	if err != nil {
		log.Fatalf("Ошибка инициализации бота: %v", err)

	}
	app := web.NewWebApp(teacherRepo, scheduleProvider)

	addr := ":8080"
	log.Printf("Запуск веб-приложения на %s...", addr)
	go app.Run(addr)
	// 6. Запускаем бота.
	telegramBot.Start()

	select {}
}
