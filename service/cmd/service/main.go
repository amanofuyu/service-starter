package main

import (
	"context"
	"log"

	"service-starter/service/internal/app"
)

func main() {
	// 进程入口只委托给 app.Run，便于启动装配逻辑集中测试和维护。
	if err := app.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
