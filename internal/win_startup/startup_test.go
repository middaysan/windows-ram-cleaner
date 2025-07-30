package winstartup

import (
	"testing"
)

func TestStartupTaskLifecycle(t *testing.T) {
	// 1. Проверяем, что задачи нет
	exists, err := IsStartupTaskExists()
	if err != nil {
		t.Fatalf("error checking startup task existence: %v", err)
	}
	if exists {
		t.Fatalf("expected startup task not to exist initially, but it exists")
	}

	// 2. Создаём задачу
	if err := CreateStartupTask(); err != nil {
		t.Fatalf("error creating startup task: %v", err)
	}

	// 3. Проверяем, что задача есть
	exists, err = IsStartupTaskExists()
	if err != nil {
		t.Fatalf("error checking startup task after creation: %v", err)
	}
	if !exists {
		t.Fatalf("expected startup task to exist after creation, but it does not")
	}

	// 4. Удаляем задачу
	if err := DeleteStartupTask(); err != nil {
		t.Fatalf("error deleting startup task: %v", err)
	}

	// 5. Проверяем, что задачи больше нет
	exists, err = IsStartupTaskExists()
	if err != nil {
		t.Fatalf("error checking startup task after deletion: %v", err)
	}
	if exists {
		t.Fatalf("expected startup task not to exist after deletion, but it exists")
	}
}
