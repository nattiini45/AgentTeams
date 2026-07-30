package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	v1beta1 "github.com/agentscope-ai/AgentTeams/agentteams-controller/api/v1beta1"
	"github.com/agentscope-ai/AgentTeams/agentteams-controller/internal/backend"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestAppendWorkerEventCapsAndOrdersNewestFirst(t *testing.T) {
	w := &v1beta1.Worker{}
	for i := 0; i < v1beta1.MaxWorkerRecentEvents+10; i++ {
		v1beta1.AppendWorkerEvent(w, v1beta1.WorkerEvent{
			Type:   "lifecycle",
			Reason: "wake",
		})
	}
	if got := len(w.Status.RecentEvents); got != v1beta1.MaxWorkerRecentEvents {
		t.Fatalf("expected %d events, got %d", v1beta1.MaxWorkerRecentEvents, got)
	}

	// Newest-first: a freshly appended event must be at index 0.
	w2 := &v1beta1.Worker{}
	v1beta1.AppendWorkerEvent(w2, v1beta1.WorkerEvent{Reason: "sleep"})
	v1beta1.AppendWorkerEvent(w2, v1beta1.WorkerEvent{Reason: "wake"})
	if w2.Status.RecentEvents[0].Reason != "wake" {
		t.Fatalf("expected newest event first, got %q", w2.Status.RecentEvents[0].Reason)
	}
	if w2.Status.RecentEvents[1].Reason != "sleep" {
		t.Fatalf("expected oldest event last, got %q", w2.Status.RecentEvents[1].Reason)
	}
}

func TestLifecycleWakeRecordsEvent(t *testing.T) {
	scheme := newLifecycleTestScheme(t)
	sleeping := "Sleeping"
	worker := &v1beta1.Worker{
		ObjectMeta: metav1.ObjectMeta{Name: "alpha-dev", Namespace: "default"},
		Spec:       v1beta1.WorkerSpec{State: &sleeping},
		Status:     v1beta1.WorkerStatus{Phase: "Sleeping"},
	}
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1beta1.Worker{}).
		WithObjects(worker).
		Build()
	backendStub := &stubWorkerBackend{status: backend.StatusRunning}
	handler := NewLifecycleHandler(k8sClient, backend.NewRegistry([]backend.WorkerBackend{backendStub}), "default", nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/workers/alpha-dev/wake", nil)
	req.SetPathValue("name", "alpha-dev")
	rec := httptest.NewRecorder()
	handler.Wake(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var updated v1beta1.Worker
	if err := k8sClient.Get(context.Background(), client.ObjectKey{Name: "alpha-dev", Namespace: "default"}, &updated); err != nil {
		t.Fatalf("get worker: %v", err)
	}
	if len(updated.Status.RecentEvents) == 0 {
		t.Fatalf("expected a lifecycle event to be recorded")
	}
	ev := updated.Status.RecentEvents[0]
	if ev.Type != "lifecycle" || ev.Reason != "wake" {
		t.Fatalf("expected lifecycle/wake event, got %s/%s", ev.Type, ev.Reason)
	}
	if ev.Timestamp == "" {
		t.Fatalf("expected event timestamp to be set")
	}
}

func TestGetWorkerEventsReturnsRecordedEvents(t *testing.T) {
	scheme := newLifecycleTestScheme(t)
	worker := &v1beta1.Worker{
		ObjectMeta: metav1.ObjectMeta{Name: "alpha-dev", Namespace: "default"},
		Status: v1beta1.WorkerStatus{
			Phase: "Running",
			RecentEvents: []v1beta1.WorkerEvent{
				{Type: "health", Reason: "health-transition", Message: "healthy -> zombie", Timestamp: "2026-07-30T00:00:00Z"},
				{Type: "lifecycle", Reason: "wake", Message: "worker woken", Timestamp: "2026-07-29T00:00:00Z"},
			},
		},
	}
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1beta1.Worker{}).
		WithObjects(worker).
		Build()
	rh := NewResourceHandler(k8sClient, "default", backend.NewRegistry(nil), "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workers/alpha-dev/events", nil)
	req.SetPathValue("name", "alpha-dev")
	rec := httptest.NewRecorder()
	rh.GetWorkerEvents(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp WorkerEventListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 2 || len(resp.Events) != 2 {
		t.Fatalf("expected 2 events, got total=%d len=%d", resp.Total, len(resp.Events))
	}
	if resp.Events[0].Reason != "health-transition" {
		t.Fatalf("expected newest event first, got %q", resp.Events[0].Reason)
	}
}

func TestGetWorkerEventsEmptyWhenNone(t *testing.T) {
	scheme := newLifecycleTestScheme(t)
	worker := &v1beta1.Worker{
		ObjectMeta: metav1.ObjectMeta{Name: "alpha-dev", Namespace: "default"},
		Status:     v1beta1.WorkerStatus{Phase: "Running"},
	}
	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithStatusSubresource(&v1beta1.Worker{}).
		WithObjects(worker).
		Build()
	rh := NewResourceHandler(k8sClient, "default", backend.NewRegistry(nil), "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/workers/alpha-dev/events", nil)
	req.SetPathValue("name", "alpha-dev")
	rec := httptest.NewRecorder()
	rh.GetWorkerEvents(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	var resp WorkerEventListResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Total != 0 || resp.Events == nil {
		t.Fatalf("expected empty (non-nil) events list, got total=%d events=%v", resp.Total, resp.Events)
	}
}
