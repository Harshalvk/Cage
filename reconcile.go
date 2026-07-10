package main

import (
	"context"
	"log"
)

func Reconcile(ctx context.Context, sm *SandboxManager, store *Store) error {
	all, err := store.List(ctx)
	if err != nil {
		return err
	}

	for _, sb := range all {
		if sb.Status != StatusRunning {
			continue
		}

		running, err := sm.IsRunning(ctx, sb.ContainerID)
		if err != nil {
			log.Printf("reconcile: failed to check sandbox %s: %v", sb.ID, err)
		}
		if !running {
			log.Printf("reconcile: sanbox %s marked running in DB but container is gone — cleaning up", sb.ID)

			if err := store.Delete(ctx, sb.ID); err != nil {
				log.Printf("reconcile: failed to delete sandbox %s: %v", sb.ID, err)
			}
		}
	}
	return nil
}
