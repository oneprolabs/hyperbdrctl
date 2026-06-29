package commands

import (
	"fmt"
	"strings"
	"time"

	appblockstorage "hyperbdr-client/internal/app/blockstorage"
	appcloudaccount "hyperbdr-client/internal/app/cloudaccount"
	appobjectstorage "hyperbdr-client/internal/app/objectstorage"
	"hyperbdr-client/internal/output"
)

func runTargetAccountWait(ctx *context, args []string) error {
	fs := newFlagSet("cloud-account wait")
	id := fs.String("id", "", "")
	intervalSeconds := fs.Int("interval-seconds", 60, "")
	timeoutSeconds := fs.Int("timeout-seconds", 3600, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	service := appcloudaccount.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.Wait(appcloudaccount.WaitSpec{
		ID:       *id,
		Interval: time.Duration(*intervalSeconds) * time.Second,
		Timeout:  time.Duration(*timeoutSeconds) * time.Second,
	})
	if err != nil {
		return err
	}
	if err := writeWaitResult(ctx, result.Rows); err != nil {
		return err
	}
	if result.Failed {
		return waitRowsError(result.Rows, "cloud-account wait failed")
	}
	return nil
}

func runTargetCloudSyncGatewayWait(ctx *context, args []string) error {
	fs := newFlagSet("target cloud-sync-gateway wait")
	id := fs.String("id", "", "")
	intervalSeconds := fs.Int("interval-seconds", 60, "")
	timeoutSeconds := fs.Int("timeout-seconds", 3600, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	service := appblockstorage.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.Wait(appblockstorage.WaitSpec{
		ID:       *id,
		Interval: time.Duration(*intervalSeconds) * time.Second,
		Timeout:  time.Duration(*timeoutSeconds) * time.Second,
	})
	if err != nil {
		return err
	}
	if err := writeWaitResult(ctx, result.Rows); err != nil {
		return err
	}
	if result.Failed {
		return waitRowsError(result.Rows, "target cloud-sync-gateway wait failed")
	}
	return nil
}

func runTargetOSSWait(ctx *context, args []string) error {
	fs := newFlagSet("target oss wait")
	id := fs.String("id", "", "")
	intervalSeconds := fs.Int("interval-seconds", 60, "")
	timeoutSeconds := fs.Int("timeout-seconds", 3600, "")
	if err := fs.Parse(args); err != nil {
		return err
	}

	service := appobjectstorage.NewService(commandAPIAdapter{ctx: ctx})
	result, err := service.Wait(appobjectstorage.WaitSpec{
		ID:       *id,
		Interval: time.Duration(*intervalSeconds) * time.Second,
		Timeout:  time.Duration(*timeoutSeconds) * time.Second,
	})
	if err != nil {
		return err
	}
	if err := writeWaitResult(ctx, result.Rows); err != nil {
		return err
	}
	if result.Failed {
		return waitRowsError(result.Rows, "target oss wait failed")
	}
	return nil
}

func writeWaitResult(ctx *context, rows []map[string]interface{}) error {
	if ctx.cfg.Output == "json" {
		return output.JSON(ctx.out, rows)
	}
	return output.Table(ctx.out, ctx.loc, rows, waitColumns())
}

func waitRowsError(rows []map[string]interface{}, fallback string) error {
	for _, row := range rows {
		if strings.EqualFold(fmt.Sprint(row["result"]), "success") {
			continue
		}
		if errText := strings.TrimSpace(fmt.Sprint(row["error"])); errText != "" {
			return fmt.Errorf("%s", errText)
		}
	}
	return fmt.Errorf("%s", fallback)
}
