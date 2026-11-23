package uxihttp

import (
	"strings"
	"testing"
)

func TestBrowserDurableJobRecoveryContract(t *testing.T) {
	for _, expected := range []string{
		"aift.uxi.pending-job.v1",
		"sessionStorage.setItem(pendingKey",
		"sessionStorage.removeItem(pendingKey)",
		"Idempotency-Key",
		"crypto.randomUUID()",
		"resumeJob(pending.sessionID,pending.jobID)",
		"Inference outcome is uncertain",
		"do not resend",
		"current.status===\"completed\"",
		"current.status===\"interrupted\"",
	} {
		if !strings.Contains(baseIndexHTML, expected) {
			t.Errorf("browser recovery contract missing %q", expected)
		}
	}
	if strings.Contains(baseIndexHTML, "localStorage.setItem(pendingKey") {
		t.Fatal("job recovery must not persist references beyond the browser session")
	}
}
