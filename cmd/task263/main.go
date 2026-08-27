// 铁路信号联锁路径冲突复核服务入口。
//
// 支持标志：
//   --addr        HTTP 监听地址（默认 :8080）
//   --db          SQLite 数据库路径（默认 task263.db）
//   --smoke-test  运行端到端自检后退出（不启动长驻服务）
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"task263-interlock/internal/httpapi"
	"task263-interlock/internal/model"
	"task263-interlock/internal/service"
	"task263-interlock/internal/store"
)

var (
	addr      = flag.String("addr", ":8080", "HTTP 监听地址")
	dbPath    = flag.String("db", "task263.db", "SQLite 数据库路径")
	smokeTest = flag.Bool("smoke-test", false, "运行自检后退出")
)

func main() {
	flag.Parse()

	db, err := store.Open(*dbPath)
	if err != nil {
		log.Fatalf("打开数据库失败: %v", err)
	}
	defer db.Close()

	svc := service.New(db)
	if *smokeTest {
		if err := runSmoke(svc, *dbPath); err != nil {
			log.Fatalf("smoke-test 失败: %v", err)
		}
		fmt.Println("smoke-test 通过：业务闭环、持久化与重启恢复均验证成功")
		return
	}

	api := httpapi.New(svc)
	srv := &http.Server{
		Addr:    *addr,
		Handler: api.Handler(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("服务启动，监听 %s，数据库 %s", *addr, *dbPath)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("HTTP 服务错误: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("收到退出信号，关闭服务…")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("关闭服务失败: %v", err)
	}
}

// runSmoke 执行端到端自检：
//  1. 构造区段/道岔/进路（两条进路竞争同一道岔）
//  2. 创建版本并挂载进路
//  3. 验证 → 预期发现道岔竞争冲突
//  4. 关闭并重开同一数据库，验证持久化与状态恢复
//  5. 修订进路消除冲突 → 重新验证 → 通过 → 创建并发布快照
func runSmoke(svc *service.Services, dbPath string) error {
	// 1. 建拓扑
	segs := []string{"seg-a", "seg-b", "seg-c", "seg-d", "seg-e"}
	for _, id := range segs {
		if _, err := svc.Segments.Create(&model.Segment{
			ID: id, Name: "区段" + id, Kind: model.SegmentPlain, LengthM: 200,
		}); err != nil {
			return fmt.Errorf("创建区段 %s: %w", id, err)
		}
	}
	sw, err := svc.Switches.Create(&model.Switch{
		ID: "sw-1", Name: "道岔1", Position: model.SwitchNormal, NormalTo: "seg-b", ReverseTo: "seg-c",
	})
	if err != nil {
		return fmt.Errorf("创建道岔: %w", err)
	}
	_ = sw

	// 2. 两条进路竞争同一道岔 sw-1 的不同位置
	routeA, err := svc.Routes.Create(&model.Route{
		ID: "route-a", Name: "直向进路", OriginSeg: "seg-a", DestSeg: "seg-b",
		PathSegs: []string{"seg-a", "seg-b"},
		Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchNormal}},
		Release:  []model.ReleaseCondition{{SegmentIDs: []string{"seg-b"}}},
	})
	if err != nil {
		return fmt.Errorf("创建进路A: %w", err)
	}
	routeB, err := svc.Routes.Create(&model.Route{
		ID: "route-b", Name: "侧向进路", OriginSeg: "seg-a", DestSeg: "seg-c",
		PathSegs: []string{"seg-a", "seg-c"},
		Switches: []model.SwitchRequirement{{SwitchID: "sw-1", Position: model.SwitchReverse}},
		Release:  []model.ReleaseCondition{{SegmentIDs: []string{"seg-c"}}},
	})
	if err != nil {
		return fmt.Errorf("创建进路B: %w", err)
	}
	_ = routeA
	_ = routeB

	// 3. 创建版本并挂载
	ver, err := svc.Versions.Create("验证批次-1")
	if err != nil {
		return fmt.Errorf("创建版本: %w", err)
	}
	for _, rid := range []string{"route-a", "route-b"} {
		if _, err := svc.Versions.AttachRoute(ver.ID, rid); err != nil {
			return fmt.Errorf("挂载进路 %s: %w", rid, err)
		}
	}

	// 4. 验证：预期发现道岔竞争冲突
	conflicts, err := svc.ValidateVersion(ver.ID)
	if err != nil {
		return fmt.Errorf("验证失败: %w", err)
	}
	hasSwitchConflict := false
	for _, c := range conflicts {
		if c.Kind == model.ConflictSwitchContention {
			hasSwitchConflict = true
			break
		}
	}
	if !hasSwitchConflict {
		return fmt.Errorf("预期发现道岔竞争冲突，实际冲突数=%d", len(conflicts))
	}

	// 5. 重启恢复验证：关闭并重开同一数据库
	if err := svc.DB.Close(); err != nil {
		return fmt.Errorf("关闭数据库: %w", err)
	}
	db2, err := store.Open(dbPath)
	if err != nil {
		return fmt.Errorf("重开数据库: %w", err)
	}
	svc2 := service.New(db2)
	defer db2.Close()
	ver2, err := svc2.Versions.Get(ver.ID)
	if err != nil {
		return fmt.Errorf("重启后读取版本: %w", err)
	}
	if ver2.State != model.VersionHasConflict {
		return fmt.Errorf("重启后版本状态错误: %s", ver2.State)
	}
	conflicts2, err := svc2.Conflicts.ListByVersion(ver.ID)
	if err != nil {
		return fmt.Errorf("重启后读取冲突: %w", err)
	}
	if len(conflicts2) == 0 {
		return fmt.Errorf("重启后冲突丢失")
	}

	// 6. 修订：将 route-b 改为走 seg-d（不再竞争 sw-1），重新验证
	//    简化：直接排除 route-b，验证应通过
	if _, err := svc2.Versions.ExcludeRoute(ver.ID, "route-b"); err != nil {
		return fmt.Errorf("排除进路B: %w", err)
	}
	conflicts3, err := svc2.ValidateVersion(ver.ID)
	if err != nil {
		return fmt.Errorf("重新验证: %w", err)
	}
	// 排除 route-b 后 route-a 独占，无冲突
	for _, c := range conflicts3 {
		if c.Kind == model.ConflictSwitchContention || c.Kind == model.ConflictSharedSegment {
			return fmt.Errorf("修订后仍存在冲突: %s", c.Detail)
		}
	}
	ver3, err := svc2.Versions.Get(ver.ID)
	if err != nil {
		return fmt.Errorf("读取修订后版本: %w", err)
	}
	if ver3.State != model.VersionReleasable {
		return fmt.Errorf("修订后版本应为 releasable，实际 %s", ver3.State)
	}

	// 7. 创建并发布快照
	snap, err := svc2.CreateSnapshot(ver.ID)
	if err != nil {
		return fmt.Errorf("创建快照: %w", err)
	}
	published, err := svc2.Snapshots.Publish(snap.ID)
	if err != nil {
		return fmt.Errorf("发布快照: %w", err)
	}
	if published.State != model.SnapshotPublished {
		return fmt.Errorf("快照未发布")
	}
	stats, err := svc2.Stats.Summarize()
	if err != nil {
		return fmt.Errorf("统计: %w", err)
	}
	if stats.PublishedSnapshot < 1 {
		return fmt.Errorf("统计中未包含已发布快照")
	}
	return nil
}
