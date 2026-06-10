package services

import (
	"context"
	"sync"
	"testing"
	"time"

	"orderflow/worker/internal/models"
)

type mockOrderRepository struct {
	mu          sync.Mutex
	statuses    map[int]string
	transitions []string
}

func newMockOrderRepository() *mockOrderRepository {
	return &mockOrderRepository{statuses: map[int]string{}}
}

func (m *mockOrderRepository) GetStatus(orderID int) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.statuses[orderID], nil
}

func (m *mockOrderRepository) AdvanceStatus(orderID int, from, to string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.statuses[orderID] != from {
		return false, nil
	}
	m.statuses[orderID] = to
	m.transitions = append(m.transitions, to)
	return true, nil
}

func newProcessorForTest(repo *mockOrderRepository, events <-chan models.OrderEvent) *Processor {
	// delays mínimos para o teste rodar rápido
	return NewProcessor(repo, events, 2, time.Millisecond, time.Millisecond, time.Millisecond)
}

func TestProcessOrderAdvancesUntilDelivered(t *testing.T) {
	repo := newMockOrderRepository()
	repo.statuses[1] = models.StatusReceived

	processor := newProcessorForTest(repo, nil)
	processor.processOrder(context.Background(), 1)

	if repo.statuses[1] != models.StatusDelivered {
		t.Fatalf("esperava status delivered, recebeu %q", repo.statuses[1])
	}

	expected := []string{models.StatusPreparing, models.StatusReady, models.StatusDelivered}
	if len(repo.transitions) != len(expected) {
		t.Fatalf("esperava %d transições, recebeu %d: %v", len(expected), len(repo.transitions), repo.transitions)
	}
	for i, status := range expected {
		if repo.transitions[i] != status {
			t.Fatalf("transição %d: esperava %q, recebeu %q", i, status, repo.transitions[i])
		}
	}
}

func TestProcessOrderIsIdempotent(t *testing.T) {
	repo := newMockOrderRepository()
	repo.statuses[1] = models.StatusReceived

	processor := newProcessorForTest(repo, nil)

	// reprocessar o mesmo evento não pode duplicar transições
	processor.processOrder(context.Background(), 1)
	processor.processOrder(context.Background(), 1)

	if len(repo.transitions) != 3 {
		t.Fatalf("esperava 3 transições após reprocessamento, recebeu %d: %v", len(repo.transitions), repo.transitions)
	}
}

func TestProcessOrderResumesFromMiddle(t *testing.T) {
	repo := newMockOrderRepository()
	// simula um pedido que ficou no meio do fluxo após um reinício
	repo.statuses[1] = models.StatusPreparing

	processor := newProcessorForTest(repo, nil)
	processor.processOrder(context.Background(), 1)

	if repo.statuses[1] != models.StatusDelivered {
		t.Fatalf("esperava status delivered, recebeu %q", repo.statuses[1])
	}

	expected := []string{models.StatusReady, models.StatusDelivered}
	if len(repo.transitions) != len(expected) {
		t.Fatalf("esperava %d transições, recebeu %d: %v", len(expected), len(repo.transitions), repo.transitions)
	}
}

func TestProcessOrderUnknownOrderDoesNothing(t *testing.T) {
	repo := newMockOrderRepository()

	processor := newProcessorForTest(repo, nil)
	processor.processOrder(context.Background(), 42)

	if len(repo.transitions) != 0 {
		t.Fatalf("não esperava transições para pedido inexistente, recebeu: %v", repo.transitions)
	}
}

func TestRunProcessesEventsFromChannel(t *testing.T) {
	repo := newMockOrderRepository()
	repo.statuses[1] = models.StatusReceived
	repo.statuses[2] = models.StatusReceived

	events := make(chan models.OrderEvent, 2)
	events <- models.OrderEvent{OrderID: 1, Event: "order_created"}
	events <- models.OrderEvent{OrderID: 2, Event: "order_created"}
	close(events)

	processor := newProcessorForTest(repo, events)
	processor.Run(context.Background())

	if repo.statuses[1] != models.StatusDelivered || repo.statuses[2] != models.StatusDelivered {
		t.Fatalf("esperava ambos os pedidos entregues, recebeu: %v", repo.statuses)
	}
}
