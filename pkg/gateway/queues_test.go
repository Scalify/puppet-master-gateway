package gateway

import (
	"bytes"
	"context"
	"sort"
	"testing"
	"time"

	"github.com/Scalify/puppet-master-gateway/pkg/api"
	internalTesting "github.com/Scalify/puppet-master-gateway/pkg/internal/testing"
)

func TestServer_ensureQueues(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test", "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	if err := s.ensureQueues(); err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	var queues = []string{api.QueueNameJobs, api.QueueNameJobResults}
	sort.Strings(q.QueuesDeclared)
	for _, name := range queues {

		if i := sort.SearchStrings(q.QueuesDeclared, name); i > len(q.QueuesDeclared) {
			t.Errorf("Queue %s was not declared", name)
		}
	}
}

func TestServer_publishNewJob(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	_, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test", "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	job, b := newTestJob(t, "asdf-1234-asdf-1234")

	if err := s.publishNewJob(job); err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	t.Logf("message sent: %v", string(b))
	t.Logf("len Messages: %v", len(q.Messages))

	if len(q.Messages) != 1 {
		t.Fatalf("Unexpected count of sent Jobs: %d", len(q.Messages))
	}

	t.Logf("message in queue: %v", string(q.Messages[0]))
	if !bytes.Equal(b, q.Messages[0]) {
		t.Fatalf("Messages are not equal: %v , %v", string(b), string(q.Messages[0]))
	}
}

func TestServerConsumeJobResults(t *testing.T) {
	q := internalTesting.NewTestQueue()
	db := internalTesting.NewTestDB()
	b, l := internalTesting.NewTestLogger()

	s, err := NewServer(db, q, l, "test", "test")
	if err != nil {
		t.Log(err)
		t.Fatal(err)
	}

	res1, b1 := newTestJobResult(t, "asdf-1234-asdf-1234")
	res2, b2 := newTestJobResult(t, "asdf-5678-asdf-5678")
	q.Messages = append(q.Messages, b1, b2)
	t.Logf("len Messages: %v", len(q.Messages))

	job1, _ := newTestJob(t, res1.UUID)
	job2, _ := newTestJob(t, res2.UUID)
	db.Jobs = []*api.Job{job1, job2}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(1*time.Second))
	s.consumeJobResults(ctx)
	internalTesting.CheckLogger(t, b)

	time.Sleep(10 * time.Millisecond)
	cancel()

	t.Logf("len Queue Messages: %v", len(q.Messages))
	t.Logf("Len Saved Messages: %v", len(db.SavedJobs))

	if len(db.SavedJobs) != 2 {
		t.Logf("Messages stored: %v", db.GetUUIDs(db.SavedJobs))
		t.Fatalf("Unexpected count of sent results: %v", len(db.SavedJobs))
	}

	if db.SavedJobs[0].UUID != res1.UUID {
		t.Logf("Messages stored: %v", db.GetUUIDs(db.SavedJobs))
		t.Errorf("Expected to find %s, got %s", res1.UUID, db.SavedJobs[0].UUID)
	}

	if db.SavedJobs[1].UUID != res2.UUID {
		t.Logf("Messages stored: %v", db.GetUUIDs(db.SavedJobs))
		t.Errorf("Expected to find %s, got %s", res2.UUID, db.SavedJobs[1].UUID)
	}
}
