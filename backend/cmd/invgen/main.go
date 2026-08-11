// Package main 邀请码批量生成 CLI
//
// 用法：
//   go run ./cmd/invgen -count 50 -note "v1 内测"
//   go run ./cmd/invgen -count 10 -note "好友邀请" -created_by 1
//
// 读取 backend/.env 中的数据库配置（与 server 共用）。
// 输出：表格打印所有生成的码到 stdout。
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"

	"github.com/soft-test-study/backend/internal/config"
	"github.com/soft-test-study/backend/internal/database"
	"github.com/soft-test-study/backend/internal/repository"
	"github.com/soft-test-study/backend/internal/service"
)

func main() {
	var (
		count     = flag.Int("count", 10, "生成数量（1-1000）")
		createdBy = flag.Uint("created_by", 0, "生成者 user_id（admin）；不指定则填 0")
		note      = flag.String("note", "", "批次备注，如 v1 内测 / 朋友邀请")
	)
	flag.Parse()

	if *count <= 0 {
		log.Fatalf("count 必须大于 0")
	}

	cfg := config.Load()
	db, err := database.Init(cfg)
	if err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}

	repo := repository.NewInvitationCodeRepo(db)
	svc := service.NewInvitationCodeService(repo)

	codes, err := svc.Generate(*count, *createdBy, *note)
	if err != nil {
		log.Fatalf("生成邀请码失败: %v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\t邀请码\t备注\t创建时间")
	fmt.Fprintln(w, "--\t------\t----\t--------")
	for _, c := range codes {
		fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", c.ID, c.Code, c.Note, c.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	w.Flush()

	fmt.Printf("\n共生成 %d 个邀请码\n", len(codes))
}
