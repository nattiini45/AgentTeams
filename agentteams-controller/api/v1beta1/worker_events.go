package v1beta1

// MaxWorkerRecentEvents caps WorkerStatus.RecentEvents. The buffer is newest
// first; appends beyond the cap drop the oldest events.
const MaxWorkerRecentEvents = 50

// AppendWorkerEvent prepends ev to the worker's RecentEvents ring buffer and
// trims the buffer to MaxWorkerRecentEvents (newest first). It mutates the
// worker in place; callers persist via Status().Update afterwards.
func AppendWorkerEvent(w *Worker, ev WorkerEvent) {
	if w == nil {
		return
	}
	w.Status.RecentEvents = append([]WorkerEvent{ev}, w.Status.RecentEvents...)
	if len(w.Status.RecentEvents) > MaxWorkerRecentEvents {
		w.Status.RecentEvents = w.Status.RecentEvents[:MaxWorkerRecentEvents]
	}
}
